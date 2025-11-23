package main

import (
	"PullRequestService/internal/app"
	"log"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}

	if err := a.E.Start(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
