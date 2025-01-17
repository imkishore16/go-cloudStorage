package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/imkishore16/go-cloudStorage/internal/handler"
	"github.com/imkishore16/go-cloudStorage/internal/repository"
	"github.com/imkishore16/go-cloudStorage/internal/service"
)

type dataSources struct {
	S3Client *s3.Client
}

func initDS() (*dataSources, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "b516f6de4eef29a1624b6363232432eb",
				SecretAccessKey: "21e6934a65a63561b7e29725b8feb310af4636fc850dc0aed5feae37d3922617",
			}, nil
		})),
		config.WithRegion("auto"),
		config.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:           "https://2064464b91a19dc546f9951951d332b1.r2.cloudflarestorage.com",
				SigningRegion: "us-east-1",
			}, nil
		})),
		config.WithClientLogMode(aws.LogRetries|aws.LogRequestWithBody|aws.LogResponseWithBody),
	)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)
	return &dataSources{
		S3Client: s3Client,
	}, nil
}

// Inject sets up dependencies and routes
func inject(d *dataSources) (*gin.Engine, error) {
	log.Println("Injecting data sources")

	bucketName := "mmworks-poc"
	imageRepository := repository.NewImageRepository(d.S3Client, bucketName)
	imageService := service.NewImageService(imageRepository)
	imageHandler := handler.NewImageHandler(imageService)
	router := gin.Default()

	router.POST("/images", func(c *gin.Context) {
		imageHandler.PostImage(c)
	})
	router.GET("/images/:id", func(c *gin.Context) {
		imageHandler.GetImage(c)
	})
	// router.DELETE("/images/:id", func(c *gin.Context) { // Delete image by ID
	// 	handlers.DeleteImage(c, imageService)
	// })

	return router, nil
}

// Main function
func main() {
	log.Println("Starting server...")

	// Initialize data sources
	ds, err := initDS()
	if err != nil {
		log.Fatalf("Unable to initialize data sources: %v\n", err)
	}

	// Inject dependencies and set up routes
	router, err := inject(ds)
	if err != nil {
		log.Fatalf("Failed to inject data sources: %v\n", err)
	}

	// Start the server
	log.Println("Server is running on port :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}

// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"
// )

// func main() {
// 	fmt.Println("Starting server...")
// 	log.Println("Starting server...")

// 	ds, err := initDS()

// 	if err != nil {
// 		log.Fatalf("Unable to initialize data sources: %v\n", err)
// 	}

// 	router, err := inject(ds)

// 	if err != nil {
// 		log.Fatalf("Failure to inject data sources: %v\n", err)
// 	}

// 	srv := &http.Server{
// 		Addr:    ":8080",
// 		Handler: router,
// 	}

// 	go func() {
// 		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			log.Fatalf("Failed to initialize server: %v\n", err)
// 		}
// 	}()

// 	log.Printf("Listening on port %v\n", srv.Addr)

// 	// Wait for kill signal of channel
// 	quit := make(chan os.Signal)

// 	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

// 	// This blocks until a signal is passed into the quit channel
// 	<-quit

// 	// The context is used to inform the server it has 5 seconds to finish
// 	// the request it is currently handling
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	// shutdown data sources
// 	// if err := ds.close(); err != nil {
// 	// 	log.Fatalf("A problem occurred gracefully shutting down data sources: %v\n", err)
// 	// }

// 	// Shutdown server
// 	log.Println("Shutting down server...")
// 	if err := srv.Shutdown(ctx); err != nil {
// 		log.Fatalf("Server forced to shutdown: %v\n", err)
// 	}
// }
