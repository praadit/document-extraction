package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	constants "textract-mongo/pkg/const"
	"textract-mongo/pkg/controller/helper"
	"textract-mongo/pkg/dto/request"
	"textract-mongo/pkg/dto/response"
	"textract-mongo/pkg/integration/chroma"
	"textract-mongo/pkg/model"
	"textract-mongo/pkg/repo"
	"textract-mongo/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/service/textract/types"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (c *Controller) ExtractDocument(ctx *gin.Context) {
	var body request.AnalyzeDocumentRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to bind request body").Error(),
		})
		return
	}

	awsS3name := "v-lab1/claim1/"
	filename := ""
	if body.File != nil && body.File.Size > 0 && body.File.Header != nil {
		file, err := body.File.Open()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to open file").Error(),
			})
			return
		}
		defer file.Close()

		ext := filepath.Ext(body.File.Filename)
		docName := utils.RandomString(16, "", "")
		filename := fmt.Sprintf("%s%s", docName, ext)
		awsS3name = awsS3name + filename

		if err := c.s3.Upload(ctx, utils.Config.AwsS3Bucket, awsS3name, file); err != nil {
			ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to upload file").Error(),
			})
			return
		}
	} else {
		awsS3name = awsS3name + body.S3ObjectName
		filename = body.S3ObjectName
	}

	// test s3
	// if err := c.s3.GetObject(ctx, utils.Config.AwsS3Bucket, awsS3name); err != nil {
	// 	ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
	// 		Message: utils.FilterError(ctx, err, "failed to access file").Error(),
	// 	})
	// 	return
	// }

	doc := model.Document{
		ID:          primitive.NewObjectID(),
		Name:        filename,
		ExtractType: body.ExtractType,
	}
	if body.ExtractType == constants.ExtractType_Text {
		output, err := c.Textract.ExtractText(ctx, nil, &utils.Config.AwsS3Bucket, &awsS3name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to extract text").Error(),
			})
			return
		}

		blocks := helper.OutputBlockToModel(output.Blocks)
		doc.Blocks = blocks
		if output.DocumentMetadata != nil {
			doc.DocumentMetadata = &model.DocumentMetadata{
				Pages: output.DocumentMetadata.Pages,
			}
		}
	} else if body.ExtractType == constants.ExtractType_Form {
		output, err := c.Textract.ExtractFormAndTable(ctx, nil, &utils.Config.AwsS3Bucket, &awsS3name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to extract text").Error(),
			})
			return
		}

		blocks := helper.OutputBlockToModel(output.Blocks)
		doc.Blocks = blocks
		if output.DocumentMetadata != nil {
			doc.DocumentMetadata = &model.DocumentMetadata{
				Pages: output.DocumentMetadata.Pages,
			}
		}
	} else {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: "extract type invalid",
		})
		return
	}

	kvSets := helper.MapFormType(doc.Blocks)
	doc.MappedKeyValue = kvSets

	tableData := helper.MapTableType(doc.Blocks)
	doc.MappedTables = tableData

	err := c.Db.InsertOne(ctx, constants.MODEL_DOCUMENT, doc)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to store data").Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response.BaseResponse{
		Data: doc,
	})
}

func (c *Controller) ExtractDocumentAsync(ctx *gin.Context) {
	var body request.AnalyzeDocumentRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to bind request body").Error(),
		})
		return
	}

	awsS3name := "v-lab1/claim1/"
	filename := ""
	if body.File != nil && body.File.Size > 0 && body.File.Header != nil {
		file, err := body.File.Open()
		if err != nil {
			ctx.JSON(http.StatusBadRequest, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to open file").Error(),
			})
			return
		}
		defer file.Close()

		ext := filepath.Ext(body.File.Filename)
		docName := utils.RandomString(16, "", "")
		filename := fmt.Sprintf("%s%s", docName, ext)
		awsS3name = awsS3name + filename

		if err := c.s3.Upload(ctx, utils.Config.AwsS3Bucket, awsS3name, file); err != nil {
			ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to upload file").Error(),
			})
			return
		}
	} else {
		awsS3name = awsS3name + body.S3ObjectName
		filename = body.S3ObjectName
	}

	doc := model.Document{
		ID:          primitive.NewObjectID(),
		Name:        filename,
		ExtractType: body.ExtractType,
	}
	err := c.Db.InsertOne(ctx, constants.MODEL_DOCUMENT, doc)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to store document").Error(),
		})
		return
	}

	job := model.Job{
		ID:         primitive.NewObjectID(),
		DocumentId: doc.ID.Hex(),
	}
	if body.ExtractType == constants.ExtractType_Text {
		output, err := c.Textract.StartExtractText(ctx, &utils.Config.AwsS3Bucket, &awsS3name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to extract text").Error(),
			})
			return
		}
		job.JobId = *output.JobId
		job.Status = string(types.JobStatusInProgress)
	} else if body.ExtractType == constants.ExtractType_Form {
		output, err := c.Textract.StartExtractFormAndTable(ctx, &utils.Config.AwsS3Bucket, &awsS3name)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to extract text").Error(),
			})
			return
		}

		job.JobId = *output.JobId
		job.Status = string(types.JobStatusInProgress)
	} else {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: "extract type invalid",
		})
		return
	}

	err = c.Db.InsertOne(ctx, constants.MODEL_JOB, job)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to store job").Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response.BaseResponse{
		Data: doc,
	})
}

func (c *Controller) SummarizeDocument(ctx *gin.Context) {
	var body request.ProcessResultRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to bind request").Error(),
		})
		return
	}

	var mQuery bson.M
	if len(body.Id) > 0 {
		id, err := primitive.ObjectIDFromHex(body.Id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to parse id").Error(),
			})
			return
		}
		mQuery = bson.M{"_id": id}
	} else if len(body.Key) > 0 {
		mQuery = bson.M{"name": body.Key}
	} else {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: "please fil the document identifier",
		})
		return
	}

	doc := model.Document{}
	err := c.Db.FindOne(ctx, constants.MODEL_DOCUMENT, mQuery, &doc)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to get document").Error(),
		})
		return
	}

	textToSummarize := ""
	for _, block := range doc.Blocks {
		if block.BlockType == string(types.BlockTypeLine) {
			textToSummarize = fmt.Sprintf("%s %s", textToSummarize, *block.Text)
		}
	}

	summary := ""
	if doc.ExtractType == constants.ExtractType_Form {
		sum, err := c.ollama.SummarizeForm(ctx, textToSummarize, doc.MappedKeyValue)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to summarize form document").Error(),
			})
			return
		}

		summary = sum
	} else if doc.ExtractType == constants.ExtractType_Text {
		sum, err := c.ollama.SummarizeText(ctx, textToSummarize)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to summarize form document").Error(),
			})
			return
		}

		summary = sum
	}

	err = c.Db.UpdateOne(ctx, constants.MODEL_DOCUMENT, mQuery, bson.D{{Key: "$set", Value: bson.D{{Key: "summary", Value: summary}}}})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to set summarize").Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response.BaseResponse{
		Data: summary,
	})
}

func (c *Controller) BedrockSummarizeDocument(ctx *gin.Context) {
	var body request.ProcessResultRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to bind request").Error(),
		})
		return
	}

	var mQuery bson.M
	if len(body.Id) > 0 {
		id, err := primitive.ObjectIDFromHex(body.Id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to parse id").Error(),
			})
			return
		}
		mQuery = bson.M{"_id": id}
	} else if len(body.Key) > 0 {
		mQuery = bson.M{"name": body.Key}
	} else {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: "please fil the document identifier",
		})
		return
	}

	doc := model.Document{}
	err := c.Db.FindOne(ctx, constants.MODEL_DOCUMENT, mQuery, &doc)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to get document").Error(),
		})
		return
	}

	textToSummarize := ""
	for _, block := range doc.Blocks {
		if block.BlockType == string(types.BlockTypeLine) {
			textToSummarize = fmt.Sprintf("%s %s", textToSummarize, *block.Text)
		}
	}

	var jsonData any = nil
	var tableData *model.TableData = nil
	if doc.ExtractType == constants.ExtractType_Form && doc.Blocks != nil && len(doc.Blocks) > 0 {
		kvSets := helper.MapFormType(doc.Blocks)
		if kvSets != nil && len(kvSets) > 0 {
			jsonData = kvSets
		}

		tableData := helper.MapTableType(doc.Blocks)
		doc.MappedTables = tableData
	}

	sum, err := c.bedrock.Summarize(ctx, textToSummarize, jsonData, tableData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to summarize form document").Error(),
		})
		return
	}

	summary := ""
	sums, err := helper.GetResponseOfBedrockConverseOutput(sum)
	if err == nil {
		summary = strings.Join(sums, ". ")
	}

	err = c.Db.UpdateOne(ctx, constants.MODEL_DOCUMENT, mQuery, bson.D{{Key: "$set", Value: bson.D{
		{Key: "summary", Value: summary},
		{Key: "bedrockResponse", Value: sum},
	}}})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to set summarize").Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response.BaseResponse{
		Data: summary,
	})
}

func (c *Controller) MapTable(ctx *gin.Context) {
	var body request.ProcessResultRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to bind request").Error(),
		})
		return
	}

	var mQuery bson.M
	if len(body.Id) > 0 {
		id, err := primitive.ObjectIDFromHex(body.Id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to parse id").Error(),
			})
			return
		}
		mQuery = bson.M{"_id": id}
	} else if len(body.Key) > 0 {
		mQuery = bson.M{"name": bson.M{"$regex": fmt.Sprintf("%s*", body.Key), "$options": ""}}
	} else {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: "please fil the document identifier",
		})
		return
	}

	docs, err := repo.Find[model.Document](ctx, c.Db, constants.MODEL_DOCUMENT, mQuery)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to get document").Error(),
		})
		return
	}

	allBlocks := []model.Block{}
	for _, doc := range docs {
		for _, block := range doc.Blocks {
			if block.Page != nil && *block.Page == 3 {
				allBlocks = append(allBlocks, block)
			}
		}
	}

	// allFormData := helper.MapFormType(allBlocks)
	// tableData := helper.MapTableType(allBlocks)
	textLayout := helper.MapLayoutTextType(allBlocks)

	ctx.JSON(http.StatusOK, response.BaseResponse{
		Data: map[string]any{
			// "formData":  allFormData,
			"text": textLayout,
			// "tableData": tableData,
		},
	})
}

func (c *Controller) EmbedDocument(ctx *gin.Context) {
	var body request.ProcessResultRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to bind request").Error(),
		})
		return
	}

	var mQuery bson.M
	if len(body.Id) > 0 {
		id, err := primitive.ObjectIDFromHex(body.Id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to parse id").Error(),
			})
			return
		}
		mQuery = bson.M{"_id": id}
	} else if len(body.Key) > 0 {
		mQuery = bson.M{"name": bson.M{"$regex": fmt.Sprintf("%s*", body.Key), "$options": ""}}
	} else {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: "please fil the document identifier",
		})
		return
	}

	docs, err := repo.Find[model.Document](ctx, c.Db, constants.MODEL_DOCUMENT, mQuery)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to get document").Error(),
		})
		return
	}

	allBlocks := []model.Block{}
	for _, doc := range docs {
		allBlocks = append(allBlocks, doc.Blocks...)
	}

	// allFormData := helper.MapFormType(allBlocks)
	chromaDocs := []chroma.ChromaDocument{}
	tableData := helper.MapTableType(allBlocks)
	if len(tableData) > 0 {
		for tableIdx, table := range tableData {
			if table.Page >= 9 && table.Page <= 12 {
				if table.TableType == string(types.EntityTypeStructuredTable) {
					for rowIdx, rowContent := range table.Data {
						rowLine := []string{}
						metadatas := map[string]any{}
						for columnIdx, columnContent := range rowContent {
							if header, ok := table.Structure[columnIdx]; ok {
								rowLine = append(rowLine, fmt.Sprintf("%s = %s", header, columnContent))
								if strings.ToLower(header) == "description" {
									pattern := `[^a-zA-Z0-9, ]`
									re := regexp.MustCompile(pattern)
									cleanedText := re.ReplaceAllString(columnContent, " ")

									metadatas[strings.ToLower(header)] = cleanedText
								}
							}
						}

						metadatas["page"] = table.Page
						metadatas["table_type"] = table.TableType
						metadatas["table_index"] = tableIdx
						metadatas["row_index"] = rowIdx

						chromaDocs = append(chromaDocs, chroma.ChromaDocument{
							Document:  fmt.Sprintf("%s.", strings.Join(rowLine, "; ")),
							Metadatas: metadatas,
						})
					}
				} else if table.TableType == string(types.EntityTypeSemiStructuredTable) {
					for rowIdx, rowContent := range table.Data {
						rowLine := []string{}
						for _, columnContent := range rowContent {
							rowLine = append(rowLine, columnContent)
						}

						chromaDocs = append(chromaDocs, chroma.ChromaDocument{
							Document: fmt.Sprintf("%s.", strings.Join(rowLine, ", ")),
							Metadatas: map[string]any{
								"page":        table.Page,
								"table_type":  table.TableType,
								"table_index": tableIdx,
								"row_index":   rowIdx,
							},
						})
					}
				}
			}
		}
	}
	textLayout := helper.MapLayoutTextType(allBlocks)
	if len(textLayout) > 0 {
		for _, text := range textLayout {
			if text.Page > 3 && text.Page <= 14 {
				fmt.Println(text.Paragraph)
				pattern := `([^a-zA-Z0-9])\.`

				re := regexp.MustCompile(pattern)
				delimitedText := re.ReplaceAllString(text.Paragraph, "$1@@@")
				splitText := strings.Split(delimitedText, "@@@")

				var splitedRes []string
				for _, str := range splitText {
					trimmed := strings.TrimSpace(str)
					if trimmed != "" {
						splitedRes = append(splitedRes, trimmed)
					}
				}

				for idx, tx := range splitedRes {
					chromaDocs = append(chromaDocs, chroma.ChromaDocument{
						Document: tx,
						Metadatas: map[string]any{
							"page":       text.Page,
							"text_index": idx,
						},
					})
				}
			}
		}
	}

	if err := c.chroma.SetCollection(ctx, "preview-fee-benchmark"); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to set chroma collection").Error(),
		})
		return
	}

	if err := c.chroma.AddRecord(ctx, chromaDocs); err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to add chroma record").Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response.BaseResponse{
		Data: chromaDocs,
	})
}

func (c *Controller) AskOllama(ctx *gin.Context) {
	var body request.AskRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to bind request").Error(),
		})
		return
	}

	if err := c.chroma.SetCollection(ctx, "preview-fee-benchmark"); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to set chroma collection").Error(),
		})
		return
	}
	queryContext, err := c.chroma.QueryContext(ctx, body.Query, 5)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to query context").Error(),
		})
		return
	}

	answer, err := c.ollama.QueryWithContext(ctx, queryContext, body.Query)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to ask query").Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, response.BaseResponse{
		Data: map[string]any{
			"query":  body.Query,
			"answer": answer,
		},
	})
}

func (c *Controller) AskBedrock(ctx *gin.Context) {
	var body request.AskRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to bind request").Error(),
		})
		return
	}

	if err := c.chroma.SetCollection(ctx, "preview-fee-benchmark"); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to set chroma collection").Error(),
		})
		return
	}
	queryContext, err := c.chroma.QueryContext(ctx, body.Query, 10)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to query context").Error(),
		})
		return
	}

	sum, err := c.bedrock.QueryWithContext(ctx, queryContext, body.Query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to summarize form document").Error(),
		})
		return
	}

	answer := ""
	sums, err := helper.GetResponseOfBedrockConverseOutput(sum)
	if err == nil {
		answer = strings.Join(sums, ". ")
	}

	ctx.JSON(http.StatusOK, response.BaseResponse{
		Data: map[string]any{
			"query":  body.Query,
			"answer": answer,
		},
	})
}

func (c *Controller) AskNoContext(ctx *gin.Context) {
	var body request.AskRequest
	if err := ctx.Bind(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to bind request").Error(),
		})
		return
	}

	// answer, err := c.ollama.QueryWithContext(ctx, "", body.Query)
	// if err != nil {
	// 	ctx.JSON(http.StatusBadRequest, response.BaseResponse{
	// 		Message: utils.FilterError(ctx, err, "failed to ask query").Error(),
	// 	})
	// 	return
	// }

	sum, err := c.bedrock.QueryWithContext(ctx, "", body.Query)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to summarize form document").Error(),
		})
		return
	}

	answer := ""
	sums, err := helper.GetResponseOfBedrockConverseOutput(sum)
	if err == nil {
		answer = strings.Join(sums, ". ")
	}

	ctx.JSON(http.StatusOK, response.BaseResponse{
		Data: map[string]any{
			"query":  body.Query,
			"answer": answer,
		},
	})
}
