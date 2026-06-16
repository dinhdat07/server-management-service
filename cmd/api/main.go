package main

import (
	"server-management-service/internal/shared/logger"

	"server-management-service/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		logger.Log.Sugar().Fatal(err)
	}

	if err := application.Run(); err != nil {
		logger.Log.Sugar().Fatal(err)
	}
}
