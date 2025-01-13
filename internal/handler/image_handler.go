package handler

import (
	"encoding/json"
	"net/http"

	"github.com/imkishore16/go-cloudStorage/internal/model/apperrors"
	"github.com/imkishore16/go-cloudStorage/internal/service"
)

type ImageHandler interface {
	UpdateImage(w http.ResponseWriter, r *http.Request)
	DeleteImage(w http.ResponseWriter, r *http.Request)
	GetImage(w http.ResponseWriter, r *http.Request)
}

type imageHandler struct {
	ImageService service.ImageService
}

// NewImageHandler initializes an ImageHandler
func NewImageHandler(imageService service.ImageService) ImageHandler {
	return &imageHandler{
		ImageService: imageService,
	}
}

// UpdateImage handles the upload or update of an image
func (h *imageHandler) UpdateImage(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("image")
	if err != nil {
		respondError(w, apperrors.NewBadRequest("invalid file upload"))
		return
	}
	defer file.Close()

	imageURL, err := h.ImageService.UpdateImage(r.Context(), file, header.Filename)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"imageURL": imageURL})
}

// DeleteImage handles the deletion of an image
func (h *imageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	imageURL := r.URL.Query().Get("imageURL")
	if imageURL == "" {
		respondError(w, apperrors.NewBadRequest("imageURL query parameter is required"))
		return
	}

	err := h.ImageService.DeleteImage(r.Context(), imageURL)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "image deleted successfully"})
}

// GetImage handles retrieval of an image URL
func (h *imageHandler) GetImage(w http.ResponseWriter, r *http.Request) {
	imageURL := r.URL.Query().Get("imageURL")
	if imageURL == "" {
		respondError(w, apperrors.NewBadRequest("imageURL query parameter is required"))
		return
	}

	url, err := h.ImageService.GetImage(r.Context(), imageURL)
	if err != nil {
		respondError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"imageURL": url})
}

// Utility Functions

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "failed to encode JSON response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(response)
}

func respondError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*apperrors.Error); ok {
		respondJSON(w, appErr.Status(), map[string]string{"error": appErr.Message})
	} else {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
