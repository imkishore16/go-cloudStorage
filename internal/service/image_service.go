package service

import (
	"context"
	"fmt"

	"github.com/imkishore16/go-cloudStorage/internal/repository"
)

// ImageService defines the interface for image operations
type ImageService interface {
	GetImage(ctx context.Context, objName string) (string, error)
	PostImage(ctx context.Context, filePath string, objectKey string) (string, error)
	UpdateImage(ctx context.Context, filePath string, objectKey string) (string, error)
	DeleteImage(ctx context.Context, objName string) error
}

// imageService is the concrete implementation of ImageService
type imageService struct {
	imageRepo repository.ImageRepository
}

// NewImageService initializes a new ImageService
func NewImageService(imageRepo repository.ImageRepository) ImageService {
	return &imageService{
		imageRepo: imageRepo,
	}
}

// GetImage retrieves an image from the bucket
func (s *imageService) GetImage(ctx context.Context, objName string) (string, error) {
	localFilePath, err := s.imageRepo.GetImage(ctx, objName)
	if err != nil {
		return "", fmt.Errorf("error in GetImage: %w", err)
	}
	return localFilePath, nil
}

// PostImage uploads a new image to the bucket
func (s *imageService) PostImage(ctx context.Context, filePath string, objectKey string) (string, error) {
	url, err := s.imageRepo.PostImage(ctx, filePath, objectKey)
	if err != nil {
		return "", fmt.Errorf("error in PostImage: %w", err)
	}
	return url, nil
}

// UpdateImage updates an existing image in the bucket
func (s *imageService) UpdateImage(ctx context.Context, filePath string, objectKey string) (string, error) {
	url, err := s.imageRepo.UpdateImage(ctx, filePath, objectKey)
	if err != nil {
		return "", fmt.Errorf("error in UpdateImage: %w", err)
	}
	return url, nil
}

// DeleteImage deletes an image from the bucket
func (s *imageService) DeleteImage(ctx context.Context, objName string) error {
	if err := s.imageRepo.DeleteImage(ctx, objName); err != nil {
		return fmt.Errorf("error in DeleteImage: %w", err)
	}
	return nil
}
