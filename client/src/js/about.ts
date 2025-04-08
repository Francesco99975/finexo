import 'katex/dist/katex.min.css'; // Import KaTeX CSS for styling
// Ignore unused imports
import katex from 'katex'; // Import KaTeX core (optional here but good to include)
import renderMathInElement from 'katex/contrib/auto-render'; // Import auto-rendering function

void katex;

// Wait for the DOM to load, then render all math formulas
// document.addEventListener('DOMContentLoaded', () => {

// });

(() => {
        renderMathInElement(document.body, {
            delimiters: [
            { left: '$$', right: '$$', display: true },    // Display math (block)
            { left: '$', right: '$', display: false },     // Inline math
            { left: '\\(', right: '\\)', display: false }, // Alternative inline math
            { left: '\\[', right: '\\]', display: true },  // Alternative display math
            ],
            throwOnError: false,  // Continue rendering even if there’s an error
        });
})()

