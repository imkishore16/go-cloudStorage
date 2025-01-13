package aws

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/upload", fileUploadHandler)

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
