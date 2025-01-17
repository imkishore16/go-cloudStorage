package repository

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type ImageRepository interface {
	GetImage(ctx context.Context, objName string) ([]byte, string, error)
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
func (r *gcImageRepository) GetImage(ctx context.Context, objName string) ([]byte, string, error) {
	output, err := r.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &r.bucketName,
		Key:    &objName,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve file: %v", err)
	}
	defer output.Body.Close()
	fmt.Println(output)

	if output.ContentType == nil {
		return nil, "", fmt.Errorf("content type is nil")
	}
	contentType := *output.ContentType
	if !isImageContentType(contentType) {
		return nil, "", fmt.Errorf("invalid file type: %s, expected image", contentType)
	}

	// Read the image data into a byte slice
	imageData, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image data: %v", err)
	}

	saveImageData(imageData, "newimageee")
	return imageData, contentType, nil
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

func saveImageData(imageData []byte, filename string) (string, error) {
	if len(imageData) == 0 {
		return "", fmt.Errorf("image data is empty")
	}

	// Create .temp directory if it doesn't exist
	tempDir := "C:/Users/imkis/OneDrive/Documents/Visual Studio 2019/go_projects/cloudStorage/temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	// Determine the image format (PNG, JPEG, etc.)
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image data: %v", err)
	}

	// Normalize the filename extension based on the format
	ext := strings.ToLower(format)
	if ext != "jpeg" && ext != "png" && ext != "gif" {
		return "", fmt.Errorf("unsupported image format: %s", ext)
	}

	if ext == "jpeg" {
		ext = "jpg"
	}

	// Generate the full file path
	filePath := filepath.Join(tempDir, fmt.Sprintf("%s.%s", filename, ext))

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Encode and save the image based on its format
	switch ext {
	case "jpg":
		if err := jpeg.Encode(file, img, nil); err != nil {
			return "", fmt.Errorf("failed to save JPEG image: %v", err)
		}
	case "png":
		if err := png.Encode(file, img); err != nil {
			return "", fmt.Errorf("failed to save PNG image: %v", err)
		}
	default:
		return "", fmt.Errorf("unsupported format: %s", ext)
	}

	return filePath, nil
}
