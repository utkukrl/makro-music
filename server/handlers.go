package main

import (
	"context"
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
