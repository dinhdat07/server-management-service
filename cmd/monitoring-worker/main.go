package main

import (
	"log"

	"server-management-service/internal/modules/monitoring/worker"
)

func main() {
	app, err := worker.NewApp()
	if err != nil {
		log.Fatalf("init monitoring worker: %v", err)
	}

	if err := app.Run(); err != nil {
		log.Fatalf("run monitoring worker: %v", err)
	}
}
