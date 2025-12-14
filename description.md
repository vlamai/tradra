**Project Title:** Perspective Trainer – Single Binary Web App

**Role:** You are a Senior Full-Stack Developer proficient in **Go (Golang)** and **Vanilla JavaScript**.

**Goal:** Create a standalone web application distributed as a single binary executable. The app helps artists train their drawing skills by analyzing hand-drawn 2-point perspective cubes.

**Architecture Overview:**
1.  **Backend (Go):**
    *   Serves the frontend (HTML/CSS/JS) using `embed`.
    *   Process raw coordinate data sent from the frontend.
    *   Performs mathematical analysis: Linear Regression for line straightness, Intersection finding for Vanishing Points.
    *   Generates visualization images (or SVG overlays) showing errors.
2.  **Frontend (HTML5 Canvas + JS):**
    *   A simple, fullscreen canvas interface.
    *   Captures stylus/mouse input using the `Pointer Events API` (specifically utilizing `getCoalescedEvents` for high precision if available).
    *   Stores strokes as arrays of coordinates `[(x,y), (x,y)...]`.
    *   Sends the stroke data (JSON) to the Go backend upon button click.

**Functional Requirements:**

**Step 1: The Drawing Interface**
*   Provide a canvas where the user draws a cube using exactly **9 strokes** (3 verticals, 3 converging left, 3 converging right).
*   Add a "Submit / Analyze" button.
*   Add a "Clear" button.
*   *Correction:* Do NOT use raster image upload. Send raw vector data (arrays of X,Y points) to the backend to ensure mathematical precision.

**Step 2: The Analysis Algorithm (Go)**
When the backend receives the 9 strokes (arrays of points):
1.  **Line Quality Analysis (Straightness):**
    *   For each stroke, calculate the "Idea Line" using **Linear Regression (Least Squares)**.
    *   Calculate the deviation (RMSE - Root Mean Square Error) of the actual points from the ideal line.
    *   Map this error to a 0-100% score (where 100% is a perfectly straight line).
2.  **Perspective Grouping (Clustering):**
    *   Automatically categorize the 9 lines into 3 groups based on their angle:
        *   Group A: Verticals (Close to 90 degrees).
        *   Group B: Lines converging to Left VP.
        *   Group C: Lines converging to Right VP.
3.  **Vanishing Point (VP) Logic:**
    *   For Group B and Group C, calculate the theoretical intersection points of the ideal lines.
    *   Since 3 lines rarely intersect at a single perfect pixel, calculate the **centroid** of their intersection points.
    *   Measure the "Convergence Error" (how far the intersections are from each other).
4.  **Output Generation:**
    *   Create a visual feedback image (or data structure):
        *   Draw the "Ideal" straight lines in **Green**.
        *   Extend the lines to the horizon to show the Vanishing Points in **Red**.
        *   Highlight lines that deviate significantly.

**Step 3: Response Display**
*   The frontend displays the analyzed image/results alongside the original drawing.
*   Show numeric scores: "Line Confidence: X%" and "Perspective Accuracy: Y%".

**Tech Stack Constraints:**
*   **Go:** Standard library + standard `image` package (or `gg` / `fogleman/gg` for easier drawing if needed). Use `embed` for static assets.
*   **JS:** No frameworks (Vanilla JS). Keep it simple.
*   **CSS:** Minimal, clean styling.

**Deliverables:**
Please write the complete code structure:
1.  `main.go` (Backend logic + Server + Math).
2.  `index.html` (Canvas UI + JS logic).
3.  Instructions on how to build the single binary.

---

### Почему это сработает:

1.  **Math Focused:** Я указал агенту использовать *Linear Regression*. Это даст тебе ту самую "честную" оценку прямых линий, о которой ты мечтал, без возни с пикселями.
2.  **Raw Data:** Мы отправляем не картинку, а координаты. Это значит, что даже если ты нарисовал тонкую линию, математика посчитает её направление идеально точно.
3.  **Single Binary:** Использование `embed` в Go сделает программу очень удобной. Скомпилировал, кинул файл в папку, запустил — работает.

**Что делать дальше:**
Скопируй этот текст на английском в сильную модель (Claude 3.5 Sonnet или GPT-4o). Они пишут код на Go лучше всего. Если возникнут вопросы по коду, который они выдадут — приноси сюда, разберем.