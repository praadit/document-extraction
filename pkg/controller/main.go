package controller

import (
	"textract-mongo/pkg/integration/aws"
	"textract-mongo/pkg/integration/chroma"
	"textract-mongo/pkg/integration/ollama"
	"textract-mongo/pkg/repo"
)

type Controller struct {
	Db       *repo.Database
	s3       *aws.S3
	Textract *aws.Textract
	bedrock  *aws.Bedrock
	ollama   *ollama.Ollama
	chroma   *chroma.Chroma
}

func NewController(db *repo.Database, s3 *aws.S3, tract *aws.Textract, bed *aws.Bedrock, ollama *ollama.Ollama, chroma *chroma.Chroma) *Controller {
	return &Controller{
		Textract: tract,
		bedrock:  bed,
		ollama:   ollama,
		s3:       s3,
		Db:       db,
		chroma:   chroma,
	}
}
