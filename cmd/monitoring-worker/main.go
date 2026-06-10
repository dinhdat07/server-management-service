package main

import (
	"log"

	"server-management-service/internal/app/monitoring"
)

func main() {
	app, err := monitoring.NewApp()
	if err != nil {
		log.Fatalf("init monitoring worker: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("run monitoring worker: %v", err)
	}
}
