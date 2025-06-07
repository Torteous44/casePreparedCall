package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	fmt.Println("Call Service starting...")
	log.Println("Call Service initialized")
	// TODO: Initialize and start the call service
}
