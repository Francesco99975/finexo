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
        // Base colors
        std: "rgb(var(--color-std) / <alpha-value>)",
        primary: "rgb(var(--color-primary) / <alpha-value>)",
        accent: "rgb(var(--color-accent) / <alpha-value>)",
        success: "rgb(var(--color-success) / <alpha-value>)",
        error: "rgb(var(--color-error) / <alpha-value>)",

        // New additional colors
        warning: "rgb(var(--color-warning) / <alpha-value>)",
        info: "rgb(var(--color-info) / <alpha-value>)",

        // Text colors
        "text-primary": "rgb(var(--color-text-primary) / <alpha-value>)",
        "text-secondary": "rgb(var(--color-text-secondary) / <alpha-value>)",

        // Background variations
        "bg-light": "rgb(var(--color-bg-light) / <alpha-value>)",
        "bg-dark": "rgb(var(--color-bg-dark) / <alpha-value>)",

        // Standard colors
        transparent: "transparent",
        current: "currentColor",
      },
      // You might want to add these useful background and text utility classes
      backgroundColor: {
        primary: "rgb(var(--color-primary) / <alpha-value>)",
        accent: "rgb(var(--color-accent) / <alpha-value>)",
        success: "rgb(var(--color-success) / <alpha-value>)",
        error: "rgb(var(--color-error) / <alpha-value>)",
        warning: "rgb(var(--color-warning) / <alpha-value>)",
        info: "rgb(var(--color-info) / <alpha-value>)",
        light: "rgb(var(--color-bg-light) / <alpha-value>)",
        dark: "rgb(var(--color-bg-dark) / <alpha-value>)",
      },
      textColor: {
        primary: "rgb(var(--color-text-primary) / <alpha-value>)",
        secondary: "rgb(var(--color-text-secondary) / <alpha-value>)",
        accent: "rgb(var(--color-accent) / <alpha-value>)",
        success: "rgb(var(--color-success) / <alpha-value>)",
        error: "rgb(var(--color-error) / <alpha-value>)",
        warning: "rgb(var(--color-warning) / <alpha-value>)",
        info: "rgb(var(--color-info) / <alpha-value>)",
      },
    },
  },
  plugins: [],
};
