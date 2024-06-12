package controller

import (
	"context"
	"fmt"
	"log"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/collection"
	"github.com/amikos-tech/chroma-go/ollama"
	"github.com/amikos-tech/chroma-go/types"
)

func InitChroma() {
	// documents := []string{
	// 	"ayam bakar",
	// 	"ayam crispy",
	// 	"tahu crispy",
	// }
	// the `/api/embeddings` endpoint is automatically appended to the base URL
	ef, err := ollama.NewOllamaEmbeddingFunction(ollama.WithBaseURL("http://127.0.0.1:11434"), ollama.WithModel("llama3"))
	if err != nil {
		fmt.Printf("Error creating Ollama embedding function: %s \n", err)
	}
	// _, err = ef.EmbedDocuments(context.Background(), documents)
	// if err != nil {
	// 	fmt.Printf("Error embedding documents: %s \n", err)
	// }
	// fmt.Printf("Embedding response: %v \n", resp)

	// Create a new Chroma client
	client, err := chroma.NewClient("http://localhost:9000")
	if err != nil {
		panic(err)
	}

	// Create a new collection with options
	newCollection, err := client.NewCollection(
		context.TODO(),
		collection.WithName("test-collection"),
		collection.WithMetadata("key1", "value1"),
		collection.WithEmbeddingFunction(ef),
		collection.WithHNSWDistanceFunction(types.L2),
	)
	if err != nil {
		log.Fatalf("Error creating collection: %s \n", err)
	}

	// Create a new record set with to hold the records to insert
	rs, err := types.NewRecordSet(
		types.WithEmbeddingFunction(ef),
		types.WithIDGenerator(types.NewULIDGenerator()),
	)
	if err != nil {
		log.Fatalf("Error creating record set: %s \n", err)
	}
	// Add a few records to the record set
	rs.WithRecord(types.WithDocument("My name is John. And I have two dogs."), types.WithMetadata("key1", "value1"))
	rs.WithRecord(types.WithDocument("My name is Jane. I am a data scientist."), types.WithMetadata("key2", "value2"))

	// Build and validate the record set (this will create embeddings if not already present)
	_, err = rs.BuildAndValidate(context.TODO())
	if err != nil {
		log.Fatalf("Error validating record set: %s \n", err)
	}

	// Add the records to the collection
	_, err = newCollection.AddRecords(context.Background(), rs)
	if err != nil {
		log.Fatalf("Error adding documents: %s \n", err)
	}

	// Count the number of documents in the collection
	countDocs, qrerr := newCollection.Count(context.TODO())
	if qrerr != nil {
		log.Fatalf("Error counting documents: %s \n", qrerr)
	}

	// Query the collection
	fmt.Printf("countDocs: %v\n", countDocs) //this should result in 2
	qr, qrerr := newCollection.Query(context.TODO(), []string{"I love science"}, 5, nil, nil, nil)
	if qrerr != nil {
		log.Fatalf("Error querying documents: %s \n", qrerr)
	}
	fmt.Printf("qr: %v\n", qr.Documents[0][0])
}
