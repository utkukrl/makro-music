package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"photo_service/photomanagementproto"

	"google.golang.org/api/option"
	"google.golang.org/api/vision/v1"
	"google.golang.org/grpc"

	_ "github.com/lib/pq"
)

const (
	dbURL = "postgres://postgres:admin123@localhost/makro?sslmode=disable"
)

var db *sql.DB
var visionClient *vision.Service

func init() {
	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}

	visionAPIKey, err := readAPIKey()
	if err != nil {
		panic(err)
	}
	visionClient, err = vision.NewService(context.Background(), option.WithAPIKey(visionAPIKey))
	if err != nil {
		panic(err)
	}
}

func main() {
	lis, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

	srv := grpc.NewServer()
	photomanagementproto.RegisterPhotoServiceServer(srv, &myPhotoService{})
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}

	page := 1
	pageSize := 10

	result, err := GetImageFeed(page, pageSize)
	if err != nil {
		panic(err)
	}

	for _, imageData := range result {
		fmt.Printf("Upload Date: %s, Joy: %f, Sorrow: %f, Anger: %f, Surprise: %f\n", imageData.UploadDate, imageData.Joy, imageData.Sorrow, imageData.Anger, imageData.Surprise)
	}

	currentImageID := len(result)

	imageDetail, err := GetImageDetail(currentImageID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Image ID: %d, Upload Date: %s, Joy: %f, Sorrow: %f, Anger: %f, Surprise: %f\n", currentImageID, imageDetail.UploadDate, imageDetail.Joy, imageDetail.Sorrow, imageDetail.Anger, imageDetail.Surprise)
}
