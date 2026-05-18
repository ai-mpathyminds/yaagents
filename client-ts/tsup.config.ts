import { defineConfig } from "tsup";

export default defineConfig({
  entry: ["src/index.ts"],
  format: ["esm", "cjs"],
  // Emit .mjs for ESM and .cjs for CJS regardless of package "type"
  outExtension({ format }) {
    return { js: format === "esm" ? ".mjs" : ".cjs" };
  },
  // Bundle one declaration file for both formats
  dts: { only: false },
  clean: true,
  // Preserve ES2022 private fields (#field syntax) — do NOT transpile down
  target: "es2022",
  sourcemap: true,
  // Zero runtime deps — mark nothing as external; bundle is self-contained
  noExternal: [],
  // Tree-shakeable ESM: no splitting needed (single entry)
  splitting: false,
  treeshake: true,
});
