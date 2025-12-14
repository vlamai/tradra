package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/color"
	"image/png"
	"log"
	"math"
	"net/http"

	"github.com/fogleman/gg"
)

//go:embed static/*
var staticFiles embed.FS

// Point represents a 2D coordinate
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Stroke represents a series of points
type Stroke []Point

// AnalysisRequest contains the strokes to analyze
type AnalysisRequest struct {
	Strokes []Stroke `json:"strokes"`
	Width   float64  `json:"width"`
	Height  float64  `json:"height"`
}

// Line represents a line in y = mx + b form
type Line struct {
	M     float64 // slope
	B     float64 // y-intercept
	Angle float64 // angle in degrees
	RMSE  float64 // root mean square error
	Score float64 // straightness score (0-100)
}

// AnalysisResult contains the analysis output
type AnalysisResult struct {
	ImageData          string       `json:"imageData"`
	LineScores         []float64    `json:"lineScores"`
	AverageLineScore   float64      `json:"averageLineScore"`
	LeftVP             *Point       `json:"leftVP"`
	RightVP            *Point       `json:"rightVP"`
	ConvergenceErrorL  float64      `json:"convergenceErrorL"`
	ConvergenceErrorR  float64      `json:"convergenceErrorR"`
	PerspectiveScore   float64      `json:"perspectiveScore"`
}

func main() {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/analyze", handleAnalyze)

	port := "8080"
	fmt.Printf("Server starting on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write(data)
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Strokes) != 9 {
		http.Error(w, "Expected exactly 9 strokes", http.StatusBadRequest)
		return
	}

	result := analyzeStrokes(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func analyzeStrokes(req AnalysisRequest) AnalysisResult {
	// Step 1: Calculate ideal lines for each stroke
	lines := make([]Line, len(req.Strokes))
	lineScores := make([]float64, len(req.Strokes))

	for i, stroke := range req.Strokes {
		lines[i] = calculateIdealLine(stroke)
		lineScores[i] = lines[i].Score
	}

	// Step 2: Cluster lines into groups (vertical, left-converging, right-converging)
	verticals, leftGroup, rightGroup := clusterLines(lines)

	// Step 3: Calculate vanishing points
	var leftVP, rightVP *Point
	var convergenceErrorL, convergenceErrorR float64

	if len(leftGroup) >= 2 {
		leftVP, convergenceErrorL = calculateVanishingPoint(lines, leftGroup)
	}
	if len(rightGroup) >= 2 {
		rightVP, convergenceErrorR = calculateVanishingPoint(lines, rightGroup)
	}

	// Step 4: Calculate perspective score
	perspectiveScore := calculatePerspectiveScore(convergenceErrorL, convergenceErrorR, req.Width, req.Height)

	// Step 5: Generate visualization
	imageData := generateVisualization(req, lines, verticals, leftGroup, rightGroup, leftVP, rightVP)

	// Calculate average line score
	avgScore := 0.0
	for _, score := range lineScores {
		avgScore += score
	}
	avgScore /= float64(len(lineScores))

	return AnalysisResult{
		ImageData:         imageData,
		LineScores:        lineScores,
		AverageLineScore:  avgScore,
		LeftVP:            leftVP,
		RightVP:           rightVP,
		ConvergenceErrorL: convergenceErrorL,
		ConvergenceErrorR: convergenceErrorR,
		PerspectiveScore:  perspectiveScore,
	}
}

// calculateIdealLine uses linear regression to find the best-fit line
func calculateIdealLine(stroke Stroke) Line {
	n := float64(len(stroke))
	if n < 2 {
		return Line{}
	}

	// Calculate means
	var sumX, sumY float64
	for _, p := range stroke {
		sumX += p.X
		sumY += p.Y
	}
	meanX := sumX / n
	meanY := sumY / n

	// Check if line is vertical (very small x variance)
	var sumXX float64
	for _, p := range stroke {
		dx := p.X - meanX
		sumXX += dx * dx
	}
	varianceX := sumXX / n

	// If nearly vertical, treat specially
	if varianceX < 1.0 {
		// Vertical line: calculate RMSE from mean X
		rmse := 0.0
		for _, p := range stroke {
			dx := p.X - meanX
			rmse += dx * dx
		}
		rmse = math.Sqrt(rmse / n)

		return Line{
			M:     math.MaxFloat64, // Infinite slope
			B:     meanX,           // Store x-position instead
			Angle: 90.0,
			RMSE:  rmse,
			Score: calculateScore(rmse),
		}
	}

	// Calculate slope and intercept using least squares
	var sumXY, sumXX2 float64
	for _, p := range stroke {
		dx := p.X - meanX
		dy := p.Y - meanY
		sumXY += dx * dy
		sumXX2 += dx * dx
	}

	m := sumXY / sumXX2
	b := meanY - m*meanX

	// Calculate RMSE
	rmse := 0.0
	for _, p := range stroke {
		predicted := m*p.X + b
		error := p.Y - predicted
		rmse += error * error
	}
	rmse = math.Sqrt(rmse / n)

	// Calculate angle
	angle := math.Atan(m) * 180.0 / math.Pi

	return Line{
		M:     m,
		B:     b,
		Angle: angle,
		RMSE:  rmse,
		Score: calculateScore(rmse),
	}
}

// calculateScore converts RMSE to a 0-100 score
func calculateScore(rmse float64) float64 {
	// Lower RMSE = higher score
	// Use exponential decay: score = 100 * e^(-rmse/threshold)
	threshold := 5.0 // Adjust based on typical canvas size
	score := 100.0 * math.Exp(-rmse/threshold)
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score
}

// clusterLines groups lines into vertical, left-converging, and right-converging
func clusterLines(lines []Line) (verticals, leftGroup, rightGroup []int) {
	for i, line := range lines {
		absAngle := math.Abs(line.Angle)

		// Vertical: angle close to 90 or -90
		if absAngle > 70 && absAngle < 110 {
			verticals = append(verticals, i)
		} else if line.Angle < 0 {
			// Negative slope: converging to right VP
			rightGroup = append(rightGroup, i)
		} else {
			// Positive slope: converging to left VP
			leftGroup = append(leftGroup, i)
		}
	}
	return
}

// calculateVanishingPoint finds the centroid of intersection points
func calculateVanishingPoint(lines []Line, group []int) (*Point, float64) {
	if len(group) < 2 {
		return nil, 0
	}

	// Find all pairwise intersections
	intersections := []Point{}
	for i := 0; i < len(group); i++ {
		for j := i + 1; j < len(group); j++ {
			line1 := lines[group[i]]
			line2 := lines[group[j]]

			intersection := findIntersection(line1, line2)
			if intersection != nil {
				intersections = append(intersections, *intersection)
			}
		}
	}

	if len(intersections) == 0 {
		return nil, 0
	}

	// Calculate centroid
	centroid := Point{}
	for _, p := range intersections {
		centroid.X += p.X
		centroid.Y += p.Y
	}
	centroid.X /= float64(len(intersections))
	centroid.Y /= float64(len(intersections))

	// Calculate convergence error (average distance from centroid)
	errorSum := 0.0
	for _, p := range intersections {
		dx := p.X - centroid.X
		dy := p.Y - centroid.Y
		errorSum += math.Sqrt(dx*dx + dy*dy)
	}
	convergenceError := errorSum / float64(len(intersections))

	return &centroid, convergenceError
}

// findIntersection finds where two lines intersect
func findIntersection(line1, line2 Line) *Point {
	// Handle vertical lines
	if line1.M == math.MaxFloat64 && line2.M == math.MaxFloat64 {
		return nil // Parallel verticals
	}
	if line1.M == math.MaxFloat64 {
		x := line1.B
		y := line2.M*x + line2.B
		return &Point{X: x, Y: y}
	}
	if line2.M == math.MaxFloat64 {
		x := line2.B
		y := line1.M*x + line1.B
		return &Point{X: x, Y: y}
	}

	// Check for parallel lines
	if math.Abs(line1.M-line2.M) < 0.001 {
		return nil
	}

	// y = m1*x + b1
	// y = m2*x + b2
	// m1*x + b1 = m2*x + b2
	// x = (b2 - b1) / (m1 - m2)
	x := (line2.B - line1.B) / (line1.M - line2.M)
	y := line1.M*x + line1.B

	return &Point{X: x, Y: y}
}

// calculatePerspectiveScore converts convergence errors to a score
func calculatePerspectiveScore(errorL, errorR, width, height float64) float64 {
	// Average the two convergence errors
	avgError := (errorL + errorR) / 2.0

	// Normalize by canvas diagonal
	diagonal := math.Sqrt(width*width + height*height)
	normalizedError := avgError / diagonal

	// Convert to 0-100 score (lower error = higher score)
	score := 100.0 * math.Exp(-normalizedError*10.0)
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score
}

// generateVisualization creates an overlay image showing the analysis
func generateVisualization(req AnalysisRequest, lines []Line, verticals, leftGroup, rightGroup []int, leftVP, rightVP *Point) string {
	width := int(req.Width)
	height := int(req.Height)

	dc := gg.NewContext(width, height)

	// Draw original strokes in light gray
	dc.SetColor(color.RGBA{200, 200, 200, 255})
	dc.SetLineWidth(2)
	for _, stroke := range req.Strokes {
		if len(stroke) == 0 {
			continue
		}
		dc.MoveTo(stroke[0].X, stroke[0].Y)
		for _, p := range stroke[1:] {
			dc.LineTo(p.X, p.Y)
		}
		dc.Stroke()
	}

	// Draw ideal lines in green
	dc.SetColor(color.RGBA{0, 200, 0, 255})
	dc.SetLineWidth(2)
	for i, stroke := range req.Strokes {
		if len(stroke) < 2 {
			continue
		}
		line := lines[i]

		// Find stroke bounds
		minX, maxX := stroke[0].X, stroke[0].X
		minY, maxY := stroke[0].Y, stroke[0].Y
		for _, p := range stroke {
			if p.X < minX {
				minX = p.X
			}
			if p.X > maxX {
				maxX = p.X
			}
			if p.Y < minY {
				minY = p.Y
			}
			if p.Y > maxY {
				maxY = p.Y
			}
		}

		if line.M == math.MaxFloat64 {
			// Vertical line
			dc.DrawLine(line.B, minY, line.B, maxY)
		} else {
			y1 := line.M*minX + line.B
			y2 := line.M*maxX + line.B
			dc.DrawLine(minX, y1, maxX, y2)
		}
		dc.Stroke()
	}

	// Extend lines to vanishing points in red
	dc.SetColor(color.RGBA{255, 0, 0, 120})
	dc.SetLineWidth(1)

	// Extend left group to left VP
	if leftVP != nil {
		for _, idx := range leftGroup {
			stroke := req.Strokes[idx]
			if len(stroke) > 0 {
				// Draw from first point to VP
				dc.DrawLine(stroke[0].X, stroke[0].Y, leftVP.X, leftVP.Y)
				dc.Stroke()
			}
		}
		// Draw VP marker
		dc.SetColor(color.RGBA{255, 0, 0, 255})
		dc.DrawCircle(leftVP.X, leftVP.Y, 8)
		dc.Fill()
	}

	// Extend right group to right VP
	dc.SetColor(color.RGBA{255, 0, 0, 120})
	if rightVP != nil {
		for _, idx := range rightGroup {
			stroke := req.Strokes[idx]
			if len(stroke) > 0 {
				dc.DrawLine(stroke[0].X, stroke[0].Y, rightVP.X, rightVP.Y)
				dc.Stroke()
			}
		}
		// Draw VP marker
		dc.SetColor(color.RGBA{255, 0, 0, 255})
		dc.DrawCircle(rightVP.X, rightVP.Y, 8)
		dc.Fill()
	}

	// Convert to base64 PNG
	var buf bytes.Buffer
	png.Encode(&buf, dc.Image())
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}
