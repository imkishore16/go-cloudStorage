package repository

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ImageRepository interface {
	GetImage(ctx context.Context, objName string) (string, error)
	PostImage(ctx context.Context, filePath string, objectKey string) (string, error)
	UpdateImage(ctx context.Context, filePath string, objectKey string) (string, error)
	DeleteImage(ctx context.Context, objectKey string) error
}

type gcImageRepository struct {
	s3Client   *s3.Client
	bucketName string
}

func NewImageRepository(s3Client *s3.Client, bucketName string) ImageRepository {
	return &gcImageRepository{
		s3Client:   s3Client,
		bucketName: bucketName,
	}
}

func (r *gcImageRepository) GetImage(ctx context.Context, objName string) (string, error) {
	output, err := r.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &r.bucketName,
		Key:    &objName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve file: %v", err)
	}
	defer output.Body.Close()

	if output.ContentType == nil {
		return "", fmt.Errorf("content type is nil")
	}

	contentType := *output.ContentType
	if !isImageContentType(contentType) {
		return "", fmt.Errorf("invalid file type: %s, expected image", contentType)
	}

	ext := getExtensionFromContentType(contentType)
	if ext == "" {
		return "", fmt.Errorf("unable to determine file extension for content type: %s", contentType)
	}

	downloadDir := "downloads"
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create downloads directory: %v", err)
	}

	localFilePath := filepath.Join(downloadDir, fmt.Sprintf("%s%s", objName, ext))
	localFile, err := os.Create(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %v", err)
	}
	defer localFile.Close()

	if _, err = io.Copy(localFile, output.Body); err != nil {
		return "", fmt.Errorf("failed to save file locally: %v", err)
	}

	return localFilePath, nil
}

func (r *gcImageRepository) PostImage(ctx context.Context, filePath string, objectKey string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file for content type detection: %v", err)
	}

	contentType := http.DetectContentType(buffer)
	file.Seek(0, io.SeekStart)

	uploader := manager.NewUploader(r.s3Client)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      &r.bucketName,
		Key:         &objectKey,
		Body:        file,
		ContentType: &contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %v", err)
	}

	return fmt.Sprintf("File '%s' uploaded successfully to bucket '%s'", objectKey, r.bucketName), nil
}

func (r *gcImageRepository) UpdateImage(ctx context.Context, filePath string, objectKey string) (string, error) {
	// Delete the existing image
	err := r.DeleteImage(ctx, objectKey)
	if err != nil {
		return "", fmt.Errorf("failed to update image - unable to delete existing file: %v", err)
	}

	// Upload the new image
	return r.PostImage(ctx, filePath, objectKey)
}

func (r *gcImageRepository) DeleteImage(ctx context.Context, objectKey string) error {
	_, err := r.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &r.bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return fmt.Errorf("failed to delete image: %v", err)
	}

	return nil
}

func isImageContentType(contentType string) bool {
	validImageTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"image/bmp":  true,
	}
	return validImageTypes[contentType]
}

func getExtensionFromContentType(contentType string) string {
	contentTypeToExt := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
		"image/bmp":  ".bmp",
	}
	return contentTypeToExt[contentType]
}
