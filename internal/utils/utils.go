package utils

var validImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

// IsAllowedImageType determines if image is among types defined
// in map of allowed images
func IsAllowedImageType(mimeType string) bool {
	_, exists := validImageTypes[mimeType]

	return exists
}
