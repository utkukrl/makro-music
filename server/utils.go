package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type APIKey struct {
	VisionAPIKey string `json:"vision_api_key"`
}

func readAPIKey() (string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	VisionApiKey := os.Getenv("VISION_API_KEY")

	file, err := os.Open(VisionApiKey)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var apiKey APIKey
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&apiKey)
	if err != nil {
		return "", err
	}

	return apiKey.VisionAPIKey, nil
}
