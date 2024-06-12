package aws

import (
	"context"
	"textract-mongo/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type AwsCredentialProvider struct{}

func (c AwsCredentialProvider) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{
		AccessKeyID:     utils.Config.AwsAccessKeyID,
		SecretAccessKey: utils.Config.AwsSecretAccessKey,
	}, nil
}
