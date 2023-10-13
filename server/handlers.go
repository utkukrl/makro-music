package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"photo_service/photomanagementproto"
	"time"

	"google.golang.org/api/vision/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type myPhotoService struct {
	photomanagementproto.UnimplementedPhotoServiceServer
}

type ImageData struct {
	ID         int
	UploadDate string
	Joy        float64
	Sorrow     float64
	Anger      float64
	Surprise   float64
}

func (s *myPhotoService) UploadImage(ctx context.Context, in *photomanagementproto.ImageRequest) (*photomanagementproto.ImageResponse, error) {
	image := &vision.Image{
		Content: base64.StdEncoding.EncodeToString(in.ImageData),
	}

	request := &vision.AnnotateImageRequest{
		Image: image,
		Features: []*vision.Feature{
			{
				Type: "FACE_DETECTION",
			},
		},
	}

	batchRequest := &vision.BatchAnnotateImagesRequest{
		Requests: []*vision.AnnotateImageRequest{request},
	}

	response, err := visionClient.Images.Annotate(batchRequest).Do()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Vision API request failed: %v", err)
	}

	if len(response.Responses) > 0 {
		labels := response.Responses[0].LabelAnnotations
		if len(labels) > 0 {
			visionResult := labels[0].Description

			imageResponse := &photomanagementproto.ImageResponse{
				Result: visionResult,
			}

			imageData := ImageData{
				UploadDate: time.Now().Format(time.RFC3339),
			}

			_, err := db.Exec("INSERT INTO image_data (upload_date, joy, sorrow, anger, surprise) VALUES ($1, $2, $3, $4, $5)",
				imageData.UploadDate, imageData.Joy, imageData.Sorrow, imageData.Anger, imageData.Surprise)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to insert data into PostgreSQL: %v", err)
			}

			return imageResponse, nil
		}
	}

	return nil, status.Errorf(codes.Internal, "No labels detected")
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
