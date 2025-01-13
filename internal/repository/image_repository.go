package repository

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/imkishore16/go-cloudStorage/internal/model/apperrors"
)

type ImageRepository interface {
	Delete(ctx context.Context, objName string) error
	Update(ctx context.Context, objName string, imageFile multipart.File) (string, error)
	Get(ctx context.Context, objName string) (string, error)
}

type gcImageRepository struct {
	Storage    *storage.Client
	BucketName string
}

// NewImageRepository is a factory for initializing User Repositories
func NewImageRepository(gcClient *storage.Client, bucketName string) ImageRepository {
	return &gcImageRepository{
		Storage:    gcClient,
		BucketName: bucketName,
	}
}

func GetImage(fileName string, s3Client *s3.Client, bucketName string) (string, error) {
	// Retrieve the file from the bucket
	output, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &fileName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve file: %v", err)
	}
	defer output.Body.Close()

	// Validate Content-Type
	if output.ContentType == nil {
		return "", fmt.Errorf("content type is nil")
	}

	contentType := *output.ContentType
	if !isImageContentType(contentType) {
		return "", fmt.Errorf("invalid file type: %s, expected image", contentType)
	}

	// Get file extension from content type
	ext := getExtensionFromContentType(contentType)
	if ext == "" {
		return "", fmt.Errorf("unable to determine file extension for content type: %s", contentType)
	}

	// Create directory if it doesn't exist
	downloadDir := "downloads"
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create downloads directory: %v", err)
	}

	// Create a local file with proper extension
	localFilePath := filepath.Join(downloadDir, fmt.Sprintf("%s%s", fileName, ext))
	localFile, err := os.Create(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %v", err)
	}
	defer localFile.Close()

	// Copy the file contents
	if _, err = io.Copy(localFile, output.Body); err != nil {
		return "", fmt.Errorf("failed to save file locally: %v", err)
	}

	return localFilePath, nil
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

func (r *gcImageRepository) Get(ctx context.Context, objName string) (string, error) {
	bckt := r.Storage.Bucket(r.BucketName)
	obj := bckt.Object(objName)

	// Check if object exists
	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrBucketNotExist || err == storage.ErrObjectNotExist {
			log.Printf("Image with name %s not found in bucket", objName)
			return "", apperrors.NewNotFound("image", objName)
		}
		log.Printf("Failed to get image attributes: %v", err)
		return "", apperrors.NewInternal()
	}

	// Or just return public URL if your bucket is public
	url := fmt.Sprintf(
		"https://storage.googleapis.com/%s/%s",
		r.BucketName,
		objName,
	)

	return url, nil
}

func (r *gcImageRepository) GetContent(ctx context.Context, objName string) ([]byte, error) {
	bckt := r.Storage.Bucket(r.BucketName)
	obj := bckt.Object(objName)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrBucketNotExist || err == storage.ErrObjectNotExist {
			log.Printf("Image with name %s not found in bucket", objName)
			return nil, apperrors.NewNotFound("image", objName)
		}
		log.Printf("Failed to get image attributes: %v", err)
		return nil, apperrors.NewInternal()
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("Failed to read file content: %v", err)
		return nil, apperrors.NewInternal()
	}

	return content, nil
}

func (r *gcImageRepository) Delete(ctx context.Context, objName string) error {
	bckt := r.Storage.Bucket(r.BucketName)

	object := bckt.Object(objName)

	if err := object.Delete(ctx); err != nil {
		log.Printf("Failed to delete image object with ID: %s from GC Storage\n", objName)
		return apperrors.NewInternal()
	}

	return nil
}

func (r *gcImageRepository) Update(
	ctx context.Context,
	objName string,
	imageFile multipart.File,
) (string, error) {
	bckt := r.Storage.Bucket(r.BucketName)

	object := bckt.Object(objName)
	wc := object.NewWriter(ctx)

	// set cache control so profile image will be served fresh by browsers
	// To do this with object handle, you'd first have to upload, then update
	wc.ObjectAttrs.CacheControl = "Cache-Control:no-cache, max-age=0"

	// multipart.File has a reader!
	if _, err := io.Copy(wc, imageFile); err != nil {
		log.Printf("Unable to write file to Google Cloud Storage: %v\n", err)
		return "", apperrors.NewInternal()
	}

	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("Writer.Close: %v", err)
	}

	imageURL := fmt.Sprintf(
		"https://storage.googleapis.com/%s/%s",
		r.BucketName,
		objName,
	)

	return imageURL, nil
}
