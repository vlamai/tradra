# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Tradra** (Perspective Trainer) is a single-binary web application for training artists in 2-point perspective drawing. Users draw cubes with 9 strokes on an HTML5 canvas, and the backend analyzes the mathematical accuracy of their perspective lines.

## Architecture

### Backend (Go)
- Embeds static assets (HTML/CSS/JS) using Go's `embed` package for single-binary distribution
- Receives raw stroke coordinate data (arrays of x,y points) from frontend
- Performs mathematical analysis:
  - **Linear Regression (Least Squares)** to calculate ideal straight lines
  - **RMSE (Root Mean Square Error)** to measure line straightness (0-100% score)
  - **Angle-based clustering** to group strokes into: Verticals (~90°), Left VP converging lines, Right VP converging lines
  - **Vanishing Point calculation** via centroid of intersection points for groups of 3 lines
  - **Convergence error** measurement for VP accuracy
- Generates visual feedback using the `image` package (or `fogleman/gg` for easier drawing)

### Frontend (Vanilla JS + HTML5 Canvas)
- Fullscreen canvas interface for drawing
- Uses **Pointer Events API** with `getCoalescedEvents()` for high-precision stylus input
- Stores each stroke as coordinate arrays: `[(x,y), (x,y), ...]`
- Sends 9 strokes (3 verticals, 3 left-converging, 3 right-converging) to backend as JSON
- Displays analyzed results: ideal lines (green), vanishing points (red), error scores

### Data Flow
1. User draws 9 strokes on canvas (vector data, not raster)
2. Frontend sends raw coordinate arrays to Go backend
3. Backend performs mathematical analysis on raw coordinates
4. Backend returns visual overlay + numeric scores
5. Frontend displays results alongside original drawing

## Development Commands

### Building
```bash
go build -o tradra main.go
```

### Running
```bash
./tradra
# Or during development:
go run main.go
```

### Testing
```bash
go test ./...
# Run specific test:
go test -run TestFunctionName
```

## Key Technical Constraints

- **No external dependencies preferred**: Use Go standard library where possible
- **Optional**: `fogleman/gg` may be used for easier image generation if needed
- **Frontend**: Pure vanilla JavaScript, no frameworks
- **Math precision**: Work with raw coordinate data, not pixel manipulation
- **Distribution**: Single executable binary with embedded assets

## Core Mathematical Components

When implementing analysis logic:
1. **Line straightness**: Linear regression on coordinate arrays, calculate RMSE deviation
2. **Angle clustering**: Group 9 lines by slope angle (vertical ~90°, left convergence, right convergence)
3. **VP intersection**: For each group of 3 lines, find pairwise intersections, calculate centroid
4. **Convergence error**: Measure distance spread of intersection points from centroid
5. **Visual output**: Overlay ideal lines, extend to horizon, mark VPs

## Expected Stroke Structure

Each stroke is an array of coordinate pairs from pointer events:
```json
{
  "strokes": [
    [[x1, y1], [x2, y2], [x3, y3], ...],
    ...
  ]
}
```

Expected 9 strokes total per submission.
