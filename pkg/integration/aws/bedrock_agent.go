package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
)

type BedrockAgent struct {
	svc *bedrockagent.Client
}

func InitBedrockAgent(aws aws.Config, region *string) *BedrockAgent {
	var svc *bedrockagent.Client
	if region != nil {
		conf, err := config.LoadDefaultConfig(context.Background(),
			config.WithCredentialsProvider(aws.Credentials),
			config.WithRegion(*region),
			func(lo *config.LoadOptions) error {
				return nil
			})
		if err != nil {
			log.Panic("failed to setup aws, err : " + err.Error())
			return nil
		}
		svc = bedrockagent.NewFromConfig(conf)
	} else {
		svc = bedrockagent.NewFromConfig(aws)
	}

	return &BedrockAgent{
		svc: svc,
	}
}
