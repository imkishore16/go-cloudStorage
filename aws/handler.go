package aws

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func isValidFileType(file []byte) bool {
	fileType := http.DetectContentType(file)
	return strings.HasPrefix(fileType, "image/") // Only allow images
}

func fileUploadHandler(w http.ResponseWriter, r *http.Request) {
	// Limit file size to 10MB. This line saves you from those accidental 100MB uploads!
	r.ParseMultipartForm(10 << 20)

	// Retrieve the file from form data
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Fprintf(w, "Uploaded File: %s\n", handler.Filename)
	fmt.Fprintf(w, "File Size: %d\n", handler.Size)
	fmt.Fprintf(w, "MIME Header: %v\n", handler.Header)

	// Now let’s save it locally
	// dst, err := createFile(handler.Filename)
	// if err != nil {
	// 	http.Error(w, "Error saving the file", http.StatusInternalServerError)
	// 	return
	// }
	// defer dst.Close()

	// Copy the uploaded file to the destination file
	// if _, err := dst.ReadFrom(file); err != nil {
	// 	http.Error(w, "Error saving the file", http.StatusInternalServerError)
	// }

	// Read the file into a byte slice to validate its type
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}

	if !isValidFileType(fileBytes) {
		http.Error(w, "Invalid file type", http.StatusUnsupportedMediaType)
		return
	}

	// Proceed with saving the file
	// if _, err := dst.Write(fileBytes); err != nil {
	// 	http.Error(w, "Error saving the file", http.StatusInternalServerError)
	// }

	if err := uploadToS3(fileBytes, handler.Filename); err != nil {
		http.Error(w, "Error uploading to S3", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "File successfully uploaded to S3!")
}
