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
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
)

type Bedrock struct {
	svc *bedrockruntime.Client
}

const (
	// modelName string = "anthropic.claude-3-sonnet-20240229-v1:0"
	modelName string = "meta.llama3-70b-instruct-v1:0"
)

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

func (b *Bedrock) Summarize(ctx context.Context, textToSummarize string, jsonData any, tableData any) (*bedrockruntime.ConverseOutput, error) {
	template := fmt.Sprintf(`Given a full text, give me a concise summary. Skip any preamble text and just give the summary. 
	<document>%s</document>`, textToSummarize)

	if jsonData != nil {
		jsonStr, _ := json.Marshal(jsonData)
		template = fmt.Sprintf(`Given a full document and form_data in json format, give me a concise summary. Skip any preamble text and just give the summary.
		<table_format>
		{
			"table_id":{
				"row_id": {
					"column_id":"value"
				}
			}
		}
		</table_format>
		<document>%s</document>
		<form_data>%s</form_data>`, textToSummarize, jsonStr)
	}

	if tableData != nil {
		tableStr, _ := json.Marshal(tableData)
		template = fmt.Sprintf(`Given a full document and table_data with table_format explained below, give me a concise summary. Skip any preamble text and just give the summary.
		<table_format>
		{
			"table_id":{
				"row_id": {
					"column_id":"value"
				}
			}
		}
		</table_format>
		<document>%s</document>
		<table_data>%s</table_data>`, textToSummarize, tableStr)
	}

	fmt.Print(template)

	output, err := b.svc.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId: aws.String(modelName),
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

func (b *Bedrock) LangchainSummarizeForm(ctx context.Context, textToSummarize string, jsonData any, tableData any) (string, error) {
	llm, err := bedrock.New(bedrock.WithClient(b.svc))
	if err != nil {
		return "", nil
	}

	template := fmt.Sprintf(`Given a full document, give me a concise summary. Skip any preamble text and just give the summary.
	<document>%s</document>`, textToSummarize)

	if jsonData != nil {
		jsonStr, _ := json.Marshal(jsonData)
		template = fmt.Sprintf(`Given a full document and form_data in json format, give me a concise summary. Skip any preamble text and just give the summary.
		<document>%s</document>
		<form_data>%s</form_data>`, textToSummarize, jsonStr)
	}

	if tableData != nil {
		tableStr, _ := json.Marshal(tableData)
		template = fmt.Sprintf(`Given a full document and table_data with table_format explained below, give me a concise summary. Skip any preamble text and just give the summary.
		<table_format>
		{
			"table_id":{
				"row_id": {
					"column_id":"value"
				}
			}
		}
		</table_format>
		<document>%s</document>
		<table_data>%s</table_data>`, textToSummarize, tableStr)
	}

	// fmt.Print(template)

	completion, err := llms.GenerateFromSinglePrompt(ctx, llm, template)
	if err != nil {
		return "", err
	}

	return completion, nil
}
func (b *Bedrock) QueryWithContext(ctx context.Context, queryContext string, query string) (*bedrockruntime.ConverseOutput, error) {
	prompt := fmt.Sprintf(`<|begin_of_text|><|start_header_id|>system<|end_header_id|>You are a helpful assistant that will help user to find the information in the given context. if you not found any, your answer should be "i dont know the answer yet"<|eot_id|>
	<|start_header_id|>user<|end_header_id|>here the context: \n\n %s \n\n here the question: \n %s \n <|eot_id|>
	<|start_header_id|>assistant<|end_header_id|>`, queryContext, query)

	output, err := b.svc.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId: aws.String(modelName),
		Messages: []types.Message{
			{
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: prompt,
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
