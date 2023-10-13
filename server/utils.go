package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type APIKey struct {
	VisionAPIKey string `json:"vision_api_key"`
}

type ImageData struct {
	ID         int
	UploadDate string
	Joy        float64
	Sorrow     float64
	Anger      float64
	Surprise   float64
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

func GetImageFeed(page int, pageSize int) ([]ImageData, error) {
	rows, err := db.Query("SELECT id, upload_date, joy, sorrow, anger, surprise FROM image_data ORDER BY upload_date DESC LIMIT $1 OFFSET $2",
		pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch data from PostgreSQL: %v", err)
	}
	defer rows.Close()

	var pagedResults []ImageData
	for rows.Next() {
		var imageData ImageData
		err := rows.Scan(&imageData.ID, &imageData.UploadDate, &imageData.Joy, &imageData.Sorrow, &imageData.Anger, &imageData.Surprise)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to scan data from PostgreSQL: %v", err)
		}
		pagedResults = append(pagedResults, imageData)
	}

	return pagedResults, nil
}

func GetImageDetail(imageID int) (*ImageData, error) {
	if imageID < 1 {
		return nil, status.Errorf(codes.NotFound, "Image not found")
	}

	row := db.QueryRow("SELECT id, upload_date, joy, sorrow, anger, surprise FROM image_data WHERE id = $1", imageID)
	var imageData ImageData
	err := row.Scan(&imageData.ID, &imageData.UploadDate, &imageData.Joy, &imageData.Sorrow, &imageData.Anger, &imageData.Surprise)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "Image not found")
		}
		return nil, status.Errorf(codes.Internal, "Failed to fetch data from PostgreSQL: %v", err)
	}

	return &imageData, nil
}

func UpdateImageDetail(imageID int, newData ImageData) error {
	if imageID < 1 {
		return status.Errorf(codes.NotFound, "Image not found")
	}

	_, err := db.Exec("UPDATE image_data SET upload_date = $1, joy = $2, sorrow = $3, anger = $4, surprise = $5 WHERE id = $6",
		newData.UploadDate, newData.Joy, newData.Sorrow, newData.Anger, newData.Surprise, imageID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to update data in PostgreSQL: %v", err)
	}

	return nil
}
