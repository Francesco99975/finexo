/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.js",
    "../views/*.html",
    "../views/*.templ",
    "../views/components/*.templ",
    "../views/icons/*.templ",
    "../views/layouts/*.templ",
  ],
  theme: {
    extend: {
      colors: {
        std: "rgb(var(--color-std) / <alpha-value>)",
        primary: "rgb(var(--color-primary) / <alpha-value>)",
        success: "rgb(var(--color-success) / <alpha-value>)",
        accent: "rgb(var(--color-accent) / <alpha-value>)",
        error: "rgb(var(--color-error) / <alpha-value>)",
        transparent: "transparent",
        current: "currentColor",
      },
    },
  },
  plugins: [],
};