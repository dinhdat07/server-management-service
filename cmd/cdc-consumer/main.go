package main

import (
	"log"

	"server-management-service/internal/infrastructure/cdc"
)

func main() {
	consumer, err := cdc.New()
	if err != nil {
		log.Fatalf("init cdc consumer: %v", err)
	}

	if err := consumer.Run(); err != nil {
		log.Fatalf("run cdc consumer: %v", err)
	}
}
