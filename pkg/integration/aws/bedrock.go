package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

type Bedrock struct {
	svc *bedrockruntime.Client
}

func InitBedrock(aws aws.Config, region *string) *Bedrock {
	var svc *bedrockruntime.Client
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
		svc = bedrockruntime.NewFromConfig(conf)
	} else {
		svc = bedrockruntime.NewFromConfig(aws)
	}

	return &Bedrock{
		svc: svc,
	}
}

type BedrockRequest struct {
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
}

type BedrockResponse struct {
	Completion string `json:"completion"`
}

func (b *Bedrock) SummarizeText(ctx context.Context, textToSummarize string) (*bedrockruntime.ConverseOutput, error) {
	template := fmt.Sprintf(`Given a full text, give me a concise summary. Skip any preamble text and just give the summary. 
	<document>%s</document>`, textToSummarize)

	fmt.Print(template)

	output, err := b.svc.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId: aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
		Messages: []types.Message{
			types.Message{
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: template,
					},
				},
				Role: types.ConversationRoleUser,
			},
		},
		InferenceConfig: &types.InferenceConfiguration{
			MaxTokens:   aws.Int32(2048),
			Temperature: aws.Float32(0.8),
		},
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (b *Bedrock) SummarizeForm(ctx context.Context, textToSummarize string, data any) (*bedrockruntime.ConverseOutput, error) {
	dataString, _ := json.Marshal(data)
	template := fmt.Sprintf(`Given a full document and form_data in json format, give me a concise summary. Skip any preamble text and just give the summary.
	<document>%s</document>
	<form_data>%s</form_data>`, textToSummarize, dataString)

	fmt.Print(template)

	output, err := b.svc.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId: aws.String("anthropic.claude-3-sonnet-20240229-v1:0"),
		Messages: []types.Message{
			{
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: template,
					},
				},
				Role: types.ConversationRoleUser,
			},
		},
		InferenceConfig: &types.InferenceConfiguration{
			MaxTokens:   aws.Int32(2048),
			Temperature: aws.Float32(0.8),
		},
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}
