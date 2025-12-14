# Perspective Trainer

A single-binary web application that helps artists train their drawing skills by analyzing hand-drawn 2-point perspective cubes.

## Features

- Interactive HTML5 canvas for drawing with high-precision stylus support
- Mathematical analysis using linear regression for line straightness
- Automatic vanishing point detection and convergence analysis
- Visual feedback with ideal lines (green) and vanishing point extensions (red)
- Real-time scoring for line quality and perspective accuracy

## Building

```bash
go build -o tradra main.go
```

This creates a single executable binary with all assets embedded.

## Running

```bash
./tradra
```

Then open your browser to `http://localhost:8080`

## How to Use

1. Draw a cube using exactly 9 strokes:
   - 3 vertical lines
   - 3 lines converging to the left vanishing point
   - 3 lines converging to the right vanishing point

2. Click "Analyze" when you have 9 strokes

3. View your results:
   - **Line Confidence**: How straight your lines are (0-100%)
   - **Perspective Accuracy**: How well your lines converge to vanishing points (0-100%)

4. The visualization shows:
   - Your original strokes (gray)
   - Ideal straight lines (green)
   - Extensions to vanishing points (red)
   - Vanishing point markers (red dots)

## Technical Details

### Backend (Go)
- Linear regression (least squares) for ideal line calculation
- RMSE (Root Mean Square Error) for straightness measurement
- Angle-based clustering to group lines into verticals and perspective lines
- Centroid calculation for vanishing points from line intersections
- Image generation using `fogleman/gg` library

### Frontend (Vanilla JavaScript)
- Pointer Events API with `getCoalescedEvents()` for high-precision input
- Stores raw coordinate data (not raster images) for mathematical precision
- Responsive canvas that fills the viewport

## Dependencies

- Go 1.16+ (for embed support)
- `github.com/fogleman/gg` (for image generation)

## License

This project is provided as-is for educational purposes.
