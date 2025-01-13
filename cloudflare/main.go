package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func post(filePath string, objectKey string, s3Client *s3.Client, bucketName string) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// fileStat, _ := file.Stat()
	buffer := make([]byte, 512) // First 512 bytes are enough to detect the content type
	file.Read(buffer)
	contentType := http.DetectContentType(buffer)

	// Reset the file pointer to the beginning
	file.Seek(0, io.SeekStart)

	// Upload the file
	uploader := manager.NewUploader(s3Client)
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         &objectKey,
		Body:        file,
		ContentType: &contentType,
	})
	if err != nil {
		log.Fatalf("failed to upload file: %v", err)
	}

	fmt.Printf("File '%s' uploaded to bucket '%s'\n", objectKey, bucketName)
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

func getAllFileNames(s3Client *s3.Client, bucketName string) {
	var continuationToken *string

	for {
		// Fetch the list of objects from the bucket
		output, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket:            &bucketName,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			log.Fatalf("failed to list objects in bucket %s: %v", bucketName, err)
		}

		// Print each object's key (file name)
		for _, object := range output.Contents {
			fmt.Println(*object.Key) // Object key (file name)
		}

		// Check if there are more objects to fetch
		if output.IsTruncated != nil && !*output.IsTruncated {
			break // No more objects to fetch
		}

		// Set the continuation token for the next batch
		continuationToken = output.NextContinuationToken
	}

	//  List objects in the bucket
	listOutput, err := s3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: &bucketName,
	})
	if err != nil {
		log.Fatalf("failed to list objects: %v", err)
	}

	fmt.Println("Objects in bucket:")
	for _, object := range listOutput.Contents {
		fmt.Printf(" - %s (Size: %d bytes)\n", *object.Key, object.Size)
	}
}

func main() {
	// Replace with your Cloudflare R2 settings
	r2Endpoint := "https://2064464b91a19dc546f9951951d332b1.r2.cloudflarestorage.com"
	accessKey := "b516f6de4eef29a1624b6363232432eb"
	secretKey := "21e6934a65a63561b7e29725b8feb310af4636fc850dc0aed5feae37d3922617"
	bucketName := "mmworks-poc"

	// Create AWS configuration for R2
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
			}, nil
		})),
		config.WithRegion("auto"), // R2 doesn't require a specific region
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           r2Endpoint,
				SigningRegion: "us-east-1", // Set this for signing compatibility
			}, nil
		})),
		config.WithClientLogMode(aws.LogRetries|aws.LogRequestWithBody|aws.LogResponseWithBody),
	)
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	// Create an S3 client for R2
	s3Client := s3.NewFromConfig(cfg)

	// -----------testing------------------

	// filePath := "C:/Users/imkis/OneDrive/Pictures/sukuna.jpg"
	// objectKey := "sukuna"
	fileName := "sukuna"

	// post(filePath, objectKey, s3Client, bucketName)
	// getAllFileNames(s3Client, bucketName)
	GetImage(fileName, s3Client, bucketName)

}
