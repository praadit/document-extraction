package controller

import (
	"context"
	"textract-mongo/pkg/integration/aws"
	"textract-mongo/pkg/integration/ollama"
	"textract-mongo/pkg/repo"
	"textract-mongo/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/config"
)

type Controller struct {
	textract *aws.Textract
	bedrock  *aws.Bedrock
	ollama   *ollama.Ollama
	db       *repo.Database
	s3       *aws.S3
}

func NewController() *Controller {
	customProvider := aws.AwsCredentialProvider{}

	awsConfig, err := config.LoadDefaultConfig(context.Background(), config.WithCredentialsProvider(customProvider), func(lo *config.LoadOptions) error {
		return nil
	})
	if err != nil {
		return nil
	}

	tract := aws.InitTextract(awsConfig)
	bed := aws.InitBedrock(awsConfig)
	s3 := aws.InitS3(awsConfig)
	ollama := ollama.InitOllama()
	db := repo.InitDatabase(utils.Config.MongoConn, utils.Config.MongoDbName)

	return &Controller{
		textract: tract,
		bedrock:  bed,
		ollama:   ollama,
		s3:       s3,
		db:       db,
	}
}
