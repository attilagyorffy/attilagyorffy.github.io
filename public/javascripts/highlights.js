// Initialize syntax highlighting if highlight.js is loaded
if (typeof hljs !== "undefined") {
    hljs.highlightAll();
}

// Draw per-line highlight rects behind <mark> elements
function drawHighlights() {
    document.querySelectorAll(".highlight-stroke").forEach(el => el.remove());
    document.querySelectorAll("mark").forEach((mark) => {
        const p = mark.closest("p");
        if (!p) return;
        if (getComputedStyle(p).position === "static") {
            p.style.position = "relative";
        }
        const rects = mark.getClientRects();
        const pRect = p.getBoundingClientRect();
        for (const r of rects) {
            const stroke = document.createElement("span");
            stroke.className = "highlight-stroke";
            stroke.style.top = (r.top - pRect.top - 1) + "px";
            stroke.style.left = (r.left - pRect.left - 5) + "px";
            stroke.style.width = (r.width + 10) + "px";
            stroke.style.height = (r.height + 2) + "px";
            p.appendChild(stroke);
        }
    });
}
// Wait for fadeIn animations to settle before measuring
setTimeout(drawHighlights, 1200);
window.addEventListener("resize", drawHighlights);
