package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

type Ollama struct {
	svc *ollama.LLM
}

const (
	model       string  = "llama3"
	temperature float64 = 0.8
)

func InitOllama() *Ollama {
	llm, err := ollama.New(ollama.WithModel(model))
	if err != nil {
		log.Fatal(err)
	}

	return &Ollama{
		svc: llm,
	}
}
func (o *Ollama) SummarizeText(ctx context.Context, textToSummarize string) (string, error) {
	template := fmt.Sprintf(`Given a full document, give me a concise summary. Skip any preamble text and just give the summary.
	<document>%s</document>`, textToSummarize)

	completion, err := o.svc.Call(ctx, template,
		llms.WithTemperature(temperature),
		// llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		// 	fmt.Print(string(chunk))
		// 	return nil
		// }),
	)
	if err != nil {
		return "", err
	}

	return completion, nil
}
func (o *Ollama) SummarizeForm(ctx context.Context, textToSummarize string, data any) (string, error) {
	dataString, _ := json.Marshal(data)
	template := fmt.Sprintf(`Given a full document and data in form of json, give me a concise summary. Skip any preamble text and just give the summary.
	<document>%s</document>
	<form_data>%s</form_data>`, textToSummarize, dataString)

	completion, err := o.svc.Call(ctx, template,
		llms.WithTemperature(temperature),
		// llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
		// 	fmt.Print(string(chunk))
		// 	return nil
		// }),
	)
	if err != nil {
		return "", err
	}

	return completion, nil
}
func (o *Ollama) ClassifyDocument(ctx context.Context, textToSummarize string) (string, error) {
	template := fmt.Sprintf(`Given a list of classes, classify the document into one of these classes. Skip any preamble text and just give the class name.
	<classes>DISCHARGE_SUMMARY, RECEIPT, PRESCRIPTION</classes>
	<document>%s<document>`, textToSummarize)

	completion, err := o.svc.Call(ctx, template,
		llms.WithTemperature(temperature),
	)
	if err != nil {
		return "", err
	}

	return completion, nil
}
