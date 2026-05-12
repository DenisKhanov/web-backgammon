import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        neo: {
          bg:     "#e0e5ec",
          light:  "#ffffff",
          dark:   "#a3b1c6",
          accent: "#6c63ff",
        },
        board: {
          green: "#2d5016",
          wood:  "#8B4513",
        },
        checker: {
          white: "#f0f0f0",
          black: "#3a3a3a",
        },
      },
      boxShadow: {
        "neo-raised": "6px 6px 12px #a3b1c6, -6px -6px 12px #ffffff",
        "neo-inset":  "inset 6px 6px 12px #a3b1c6, inset -6px -6px 12px #ffffff",
        "neo-sm":     "3px 3px 6px #a3b1c6, -3px -3px 6px #ffffff",
      },
    },
  },
  plugins: [],
};

export default config;
