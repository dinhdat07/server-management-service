package main

import (
	"server-management-service/internal/shared/logger"

	"server-management-service/internal/app/monitoring"
)

func main() {
	app, err := monitoring.NewApp()
	if err != nil {
		logger.Log.Sugar().Fatalf("init monitoring worker: %v", err)
	}

	if err := app.Run(); err != nil {
		logger.Log.Sugar().Fatalf("run monitoring worker: %v", err)
	}
}
