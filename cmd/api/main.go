package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Stefanuswilfrid/course-backend/internal/config"

	"github.com/joho/godotenv"
)

// go run ./cmd/api
func main() {
	err := godotenv.Load()
	apiEnv := os.Getenv("ENV")
	if err != nil && apiEnv == "" {
		log.Println("fail to load env", err)
	}
	config.LoadEnv()

	fmt.Println("Hello, world.")

}
