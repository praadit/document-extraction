package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type Bedrock struct {
	svc *bedrockruntime.Client
}

func InitBedrock(aws aws.Config) *Bedrock {
	svc := bedrockruntime.NewFromConfig(aws)

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

func (b *Bedrock) SummarizeDocument(ctx context.Context, textTosummarize string) (any, error) {
	template := fmt.Sprintf(`Given a full text, give me a concise summary. Skip any preamble text and just give the summary. "%s"`, textTosummarize)

	payload := BedrockRequest{Prompt: template, MaxTokensToSample: 2048}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	output, err := b.svc.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		Body:        payloadBytes,
		ModelId:     aws.String("anthropic.claude-v2"),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return nil, err
	}

	return output, nil
}
