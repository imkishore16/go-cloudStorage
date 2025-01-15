package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/imkishore16/go-cloudStorage/internal/handler"
	"github.com/imkishore16/go-cloudStorage/internal/repository"
	"github.com/imkishore16/go-cloudStorage/internal/service"
)

// will initialize a handler starting from data sources
// which inject into repository layer
// which inject into service layer
// which inject into handler layer
func inject(d *dataSources) (*gin.Engine, error) {
	log.Println("Injecting data sources")

	/*
	 * repository layer
	 */
	bucketName := "mmworks-poc"
	imageRepository := repository.NewImageRepository(d.S3Client, bucketName)

	/*
	 * service layer
	 */

	imageService := service.NewImageService(imageRepository)

	router := gin.Default()

	handler.NewImageHandler(imageService)

	return router, nil
}
