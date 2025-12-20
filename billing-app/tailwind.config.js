/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: "class",
  content: ["./templates/**/*.html", "./internal/**/*.go"],
  theme: {
    extend: {
      colors: {
        brand: { DEFAULT: "#2563eb", foreground: "#ffffff", dark: "#1d4ed8" },
      },
      boxShadow: {
        card: "0 1px 2px rgba(0,0,0,.06),0 1px 3px rgba(0,0,0,.1)",
      },
      borderRadius: { xl: "1rem" },
    },
  },
  daisyui: {
    themes: [
      {
        light: {
          ...require("daisyui/src/theming/themes")["light"],
          primary: "#6366f1",
          "primary-focus": "#4f46e5",
          accent: "#f472b6",
        },
      },
      {
        dark: {
          ...require("daisyui/src/theming/themes")["dark"],
          primary: "#6366f1",
          accent: "#f472b6",
        },
      },
    ],
  },
  plugins: [require("@tailwindcss/typography"), require("daisyui")],
};
