package aws

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	svc *s3.Client
}

func InitS3(aws aws.Config) *S3 {
	svc := s3.NewFromConfig(aws)

	return &S3{
		svc: svc,
	}
}

func (s *S3) Upload(ctx context.Context, bucketName string, objectKey string, file io.Reader) error {
	_, err := s.svc.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		log.Printf("Couldn't upload file to %v:%v. Here's why: %v\n", bucketName, objectKey, err)
	}

	return err
}

func (s *S3) GetObject(ctx context.Context, bucketName string, objectKey string) error {
	obj, err := s.svc.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &objectKey,
	})
	if err != nil {
		return err
	}

	fmt.Println(obj.ChecksumSHA256)

	return nil
}
