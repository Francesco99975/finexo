// vite.config.js

import fg from "fast-glob";
import { resolve } from "path";

/** @type {import('vite').UserConfig} */
export default {
  base: "/assets/dist",
  build: {
    outDir: "../static/dist",
    rollupOptions: {
      input: fg
        .sync(["src/*.ts", "src/js/**/*.ts"])
        .map((file) => resolve(__dirname, file)),
      output: {
        entryFileNames: "[name].js",
        assetFileNames: "[name].[ext]",
      },
    },
  },
  plugins: [],
};
