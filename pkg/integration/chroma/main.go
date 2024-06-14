package chroma

import (
	"context"
	"fmt"
	"log"
	"strings"

	chromago "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/collection"
	"github.com/amikos-tech/chroma-go/ollama"
	"github.com/amikos-tech/chroma-go/types"
)

const (
	model string = "llama3"
)

type Chroma struct {
	svc         *chromago.Client
	ollamaEmbed *ollama.OllamaEmbeddingFunction
	colls       *chromago.Collection
}

type ChromaDocument struct {
	Document  string
	MetaKey   string
	MetaValue any
}

func InitChroma(chromaAddress string, llmAddress string) *Chroma {
	client, err := chromago.NewClient(chromaAddress)
	if err != nil {
		log.Panic("failed to init ollama, err :" + err.Error())
	}

	ef, err := ollama.NewOllamaEmbeddingFunction(ollama.WithBaseURL(llmAddress), ollama.WithModel(model))
	if err != nil {
		log.Panic("failed to init ollama embeding to chroma, err :" + err.Error())
	}

	return &Chroma{
		svc:         client,
		ollamaEmbed: ef,
	}
}

func (c *Chroma) SetCollection(ctx context.Context, collectionName string) error {
	if c.colls != nil {
		return nil
	}
	coll, err := c.svc.GetCollection(ctx, collectionName, c.ollamaEmbed)
	if err != nil {
		newCollection, err := c.svc.NewCollection(
			ctx,
			collection.WithName(collectionName),
			collection.WithEmbeddingFunction(c.ollamaEmbed),
			collection.WithHNSWDistanceFunction(types.L2),
		)
		if err != nil {
			return err
		}
		coll = newCollection
	}

	c.colls = coll

	return nil
}

func (c *Chroma) AddRecord(ctx context.Context, documents []ChromaDocument) error {
	if c.colls == nil {
		return fmt.Errorf("collection are not set")
	}
	rs, err := types.NewRecordSet(
		types.WithEmbeddingFunction(c.ollamaEmbed),
		types.WithIDGenerator(types.NewULIDGenerator()),
	)
	if err != nil {
		return err
	}

	for _, doc := range documents {
		rs.WithRecord(types.WithDocument(doc.Document), types.WithMetadata(doc.MetaKey, doc.MetaValue))
	}

	_, err = rs.BuildAndValidate(ctx)
	if err != nil {
		return err
	}

	_, err = c.colls.AddRecords(ctx, rs)
	if err != nil {
		return err
	}

	return nil
}

func (c *Chroma) QueryContext(ctx context.Context, query string, resCount int32) (string, error) {
	if c.colls == nil {
		return "", fmt.Errorf("collection are not set")
	}
	embedQuery, err := c.ollamaEmbed.EmbedQuery(ctx, query)
	if err != nil {
		return "", err
	}

	qr, err := c.colls.Query(ctx, []string{embedQuery.String()}, resCount, nil, nil, nil)
	if err != nil {
		return "", err
	}

	queryContext := ""
	for _, d := range qr.Documents {
		queryContext = fmt.Sprintf("%s %s", queryContext, strings.Join(d, " "))
	}

	return queryContext, nil
}
