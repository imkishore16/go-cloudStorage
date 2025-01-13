package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/imkishore16/go-cloudStorage/internal/model/apperrors"
	"github.com/imkishore16/go-cloudStorage/internal/repository"
)

// ImageService defines methods the handler layer expects
type ImageService interface {
	UpdateImage(ctx context.Context, file multipart.File, filename string) (string, error)
	DeleteImage(ctx context.Context, imageURL string) error
	GetImage(ctx context.Context, imageURL string) (string, error)
}

// imageService implements ImageService
type imageService struct {
	ImageRepository repository.ImageRepository
}

// NewImageService is a factory for initializing Image Services
func NewImageService(imageRepository repository.ImageRepository) ImageService {
	return &imageService{
		ImageRepository: imageRepository,
	}
}

// extractObjectName extracts object name from GC Storage URL
func extractObjectName(url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("image URL is empty")
	}

	// Extract the object name from the URL
	// URL format: https://storage.googleapis.com/BUCKET_NAME/OBJECT_NAME
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid image URL format")
	}

	return parts[len(parts)-1], nil
}

// UpdateImage uploads or updates image in GC Storage
func (s *imageService) UpdateImage(
	ctx context.Context,
	file multipart.File,
	filename string,
) (string, error) {
	if file == nil {
		return "", apperrors.NewBadRequest("file cannot be nil")
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	objName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// Upload to cloud storage
	imageURL, err := s.ImageRepository.Update(ctx, objName, file)
	if err != nil {
		return "", fmt.Errorf("unable to update image: %w", err)
	}

	return imageURL, nil
}

// DeleteImage removes image from GC Storage
func (s *imageService) DeleteImage(ctx context.Context, imageURL string) error {
	if imageURL == "" {
		return apperrors.NewBadRequest("image URL cannot be empty")
	}

	objName, err := extractObjectName(imageURL)
	if err != nil {
		return apperrors.NewBadRequest(fmt.Sprintf("invalid image URL: %v", err))
	}

	if err := s.ImageRepository.Delete(ctx, objName); err != nil {
		return fmt.Errorf("unable to delete image: %w", err)
	}

	return nil
}

// GetImage retrieves image from GC Storage
func (s *imageService) GetImage(ctx context.Context, imageURL string) (string, error) {
	if imageURL == "" {
		return "", apperrors.NewBadRequest("image URL cannot be empty")
	}

	objName, err := extractObjectName(imageURL)
	if err != nil {
		return "", apperrors.NewBadRequest(fmt.Sprintf("invalid image URL: %v", err))
	}

	url, err := s.ImageRepository.Get(ctx, objName)
	if err != nil {
		return "", fmt.Errorf("unable to get image: %w", err)
	}

	return url, nil
}
