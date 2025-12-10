package main

import (
	"log"
)

func main() {
	if err := Run(); err != nil {
		log.Fatalf("Failed to start YS Cloud: %v", err)
	}
}