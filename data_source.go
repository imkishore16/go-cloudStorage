package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/lib/pq"
)

type dataSources struct {
	// DB            *sqlx.DB
	// RedisClient   *redis.Client
	S3Client *s3.Client
}

// InitDS establishes connections to fields in dataSources
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
		log.Fatalf("failed to load configuration, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	return &dataSources{
		S3Client: s3Client,
	}, nil
}

// // close to be used in graceful server shutdown
// func (d *dataSources) close() error {
// 	// if err := d.DB.Close(); err != nil {
// 	// 	return fmt.Errorf("error closing Postgresql: %w", err)
// 	// }

// 	// if err := d.RedisClient.Close(); err != nil {
// 	// 	return fmt.Errorf("error closing Redis Client: %w", err)
// 	// }

// 	if err := d.S3Client.Options().; err != nil {
// 		return fmt.Errorf("error closing Cloud Storage client: %w", err)
// 	}

// 	return nil
// }
