import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        neo: {
          bg: "#e0e5ec",
          light: "#ffffff",
          dark: "#a3b1c6",
          accent: "#6c63ff"
        }
      }
    }
  },
  plugins: []
};
export default config;
