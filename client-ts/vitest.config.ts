import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    environment: "node",
    include: ["tests/**/*.test.ts"],
    coverage: {
      provider: "v8",
      include: ["src/**/*.ts"],
      // Type-only files have no runtime code; exclude to avoid skewing numbers
      exclude: ["src/types.ts", "src/result.ts"],
      thresholds: {
        lines: 80,
        functions: 80,
        branches: 80,
        statements: 80,
      },
      reporter: ["text", "lcov"],
    },
  },
});
