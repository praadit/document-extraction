package chroma

// cheatsheet

// func InitChromaWithContext() {
// 	ctx := context.Background()
// 	documents := []string{
// 		"aditya tegar is the best backend engineer and the most handsome backend eng in bandung timur",
// 	}

// 	// the `/api/embeddings` endpoint is automatically appended to the base URL
// 	ef, err := ollama.NewOllamaEmbeddingFunction(ollama.WithBaseURL("http://127.0.0.1:11434"), ollama.WithModel("llama3"))
// 	if err != nil {
// 		fmt.Printf("Error creating Ollama embedding function: %s \n", err)
// 	}

// 	// // Create a new Chroma client
// 	client, err := chroma.NewClient("http://localhost:9000")
// 	if err != nil {
// 		panic(err)
// 	}

// 	coll, err := client.GetCollection(ctx, "llama3-embed-poc", ef)
// 	if err != nil {
// 		newCollection, err := client.NewCollection(
// 			ctx,
// 			collection.WithName("llama3-embed-poc"),
// 			collection.WithEmbeddingFunction(ef),
// 		)
// 		if err != nil {
// 			log.Fatalf("Error creating collection: %s \n", err)
// 		}
// 		coll = newCollection
// 	}

// 	rs, err := types.NewRecordSet(
// 		types.WithEmbeddingFunction(ef),
// 		types.WithIDGenerator(types.NewULIDGenerator()),
// 	)
// 	if err != nil {
// 		log.Fatalf("Error creating record set: %s \n", err)
// 	}
// 	// // Add a few records to the record set
// 	for _, doc := range documents {
// 		rs.WithRecord(types.WithDocument(doc))
// 	}

// 	// // Build and validate the record set (this will create embeddings if not already present)
// 	_, err = rs.BuildAndValidate(ctx)
// 	if err != nil {
// 		log.Fatalf("Error validating record set: %s \n", err)
// 	}

// 	// // Add the records to the collection
// 	// _, err = coll.AddRecords(queryContext.Background(), rs)
// 	// if err != nil {
// 	// 	log.Fatalf("Error adding documents: %s \n", err)
// 	// }

// 	// // Query the collection
// 	query := "who is aditya tegar?"
// 	embedQuery, err := ef.EmbedQuery(ctx, query)
// 	if err != nil {
// 		log.Fatalf("Error embed queryy: %s \n", err)
// 	}

// 	qr, qrerr := coll.Query(ctx, []string{embedQuery.String()}, 5, nil, nil, nil)
// 	if qrerr != nil {
// 		log.Fatalf("Error querying documents: %s \n", qrerr)
// 	}
// 	// fmt.Printf("qr: %v\n", qr.Documents[0][0])

// 	docs := qr.Documents

// 	queryContext := ""
// 	for _, d := range docs {
// 		queryContext = fmt.Sprintf("%s %s", queryContext, strings.Join(d, " "))
// 	}

// 	llm, err := llama.New(llama.WithModel("llama3"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	prompt := fmt.Sprintf("Context: %s\n\n Questing: %s\n\n Answer: ", queryContext, query)
// 	fmt.Println("\n\n With Context:")
// 	_, err = llm.Call(ctx, prompt,
// 		llms.WithTemperature(0.8),
// 		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
// 			fmt.Print(string(chunk))
// 			return nil
// 		}),
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func InitChromaNoContext() {
// 	ctx := context.Background()
// 	query := "who is aditya tegar?"

// 	queryContext := ""

// 	llm, err := llama.New(llama.WithModel("llama3"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	prompt := fmt.Sprintf("Context: %s\n\n Questing: %s\n\n Answer: ", queryContext, query)
// 	fmt.Println("No Context:")
// 	_, err = llm.Call(ctx, prompt,
// 		llms.WithTemperature(0.8),
// 		llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
// 			fmt.Print(string(chunk))
// 			return nil
// 		}),
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
