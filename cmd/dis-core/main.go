package main

import (
	"log"

	"dis-core/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}
