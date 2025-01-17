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
	objectKey := c.Param("id")

	localFilePath, contentType, err := h.imageService.GetImage(c.Request.Context(), objectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Image retrieved successfully",
		"localFilePath": localFilePath,
		"contentType":   contentType,
	})
}

func (h *ImageHandler) PostImage(c *gin.Context) {
	type ImageUploadRequest struct {
		ObjectKey string `json:"objectKey" binding:"required"`
		FilePath  string `json:"filePath" binding:"required"`
	}

	var req ImageUploadRequest

	// Parse JSON payload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// Call the service to upload the file using the provided filePath and objectKey
	url, err := h.imageService.PostImage(c.Request.Context(), req.FilePath, req.ObjectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image: " + err.Error()})
		return
	}

	// Respond with the uploaded file URL
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
