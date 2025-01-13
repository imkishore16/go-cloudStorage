package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/imkishore16/go-cloudStorage/internal/service"
)

type ImageHandler struct {
	imageService service.ImageService
}

func NewImageHandler(imageService service.ImageService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

func (h *ImageHandler) GetImage(c *gin.Context) {
	objName := c.Param("objName")

	localFilePath, err := h.imageService.GetImage(c.Request.Context(), objName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Image retrieved successfully",
		"localFilePath": localFilePath,
	})
}

func (h *ImageHandler) PostImage(c *gin.Context) {
	objectKey := c.PostForm("objectKey")

	// Retrieve the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file"})
		return
	}

	// Save the file temporarily
	tempFilePath := "./temp/" + file.Filename
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
		return
	}

	defer func() {
		// Clean up the temporary file
		_ = os.Remove(tempFilePath)
	}()

	// Call the service to upload the file
	url, err := h.imageService.PostImage(c.Request.Context(), tempFilePath, objectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image uploaded successfully",
		"url":     url,
	})
}

func (h *ImageHandler) UpdateImage(c *gin.Context) {
	objectKey := c.PostForm("objectKey")

	// Retrieve the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file"})
		return
	}

	// Save the file temporarily
	tempFilePath := "./temp/" + file.Filename
	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
		return
	}

	defer func() {
		// Clean up the temporary file
		_ = os.Remove(tempFilePath)
	}()

	url, err := h.imageService.UpdateImage(c.Request.Context(), tempFilePath, objectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image updated successfully",
		"url":     url,
	})
}

func (h *ImageHandler) DeleteImage(c *gin.Context) {
	objName := c.Param("objName")

	err := h.imageService.DeleteImage(c.Request.Context(), objName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image deleted successfully",
	})
}
