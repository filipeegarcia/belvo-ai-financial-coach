package main

import (
	"ai-financial-coach/internal/app"
	"os"
)

func main() {
	// Only set metrics defaults if not already set by environment
	if os.Getenv("ENABLE_METRICS") == "" {
		os.Setenv("ENABLE_METRICS", "false")
	}
	if os.Getenv("METRICS_PORT") == "" {
		os.Setenv("METRICS_PORT", "0")
	}
	if os.Getenv("GOFR_METRICS_PORT") == "" {
		os.Setenv("GOFR_METRICS_PORT", "0")
	}
	if os.Getenv("GOFR_ENABLE_METRICS") == "" {
		os.Setenv("GOFR_ENABLE_METRICS", "false")
	}
	if os.Getenv("GOFR_TELEMETRY") == "" {
		os.Setenv("GOFR_TELEMETRY", "false")
	}

	// Set production port if specified
	if port := os.Getenv("PORT"); port != "" {
		os.Setenv("GOFR_HTTP_PORT", port)
	}

	appInstance := app.CreateApp()

	app.RegisterRoutes(appInstance)

	appInstance.Run()
}
