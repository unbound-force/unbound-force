import { defineConfig } from "vitest/config"

// Coverage thresholds enforce the review-council-multi-model-fanout
// measurable-verification gate (design D14): 90 percent statements and
// 85 percent branches. The plugin runtime is provider, GitHub, Dewey,
// and network free, so every test runs in the Node environment with the
// v8 coverage provider.
export default defineConfig({
  test: {
    environment: "node",
    include: ["test/**/*.test.ts"],
    coverage: {
      provider: "v8",
      include: [
        "plugins/**/*.ts",
        "lib/**/*.ts",
      ],
      exclude: [
        "plugins/**/*.test.ts",
        "lib/**/*.test.ts",
        "test/**",
      ],
      thresholds: {
        statements: 90,
        branches: 85,
      },
      reporter: ["text", "text-summary"],
    },
  },
})
