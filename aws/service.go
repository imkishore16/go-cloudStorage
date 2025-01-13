package aws

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func uploadToS3(file []byte, filename string) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1"), // Your AWS region
	})
	if err != nil {
		return err
	}

	s3Client := s3.New(sess)
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("your-bucket-name"),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(file),
		ACL:    aws.String("public-read"),
	})
	return err
}
