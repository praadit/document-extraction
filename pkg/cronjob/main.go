package cronjob

import (
	"context"
	"fmt"
	constants "textract-mongo/pkg/const"
	"textract-mongo/pkg/controller"
	"textract-mongo/pkg/controller/helper"
	"textract-mongo/pkg/model"
	"textract-mongo/pkg/repo"
	"textract-mongo/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/service/textract/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Cron struct {
	control *controller.Controller
}

func InitCron(control *controller.Controller) *Cron {
	return &Cron{
		control: control,
	}
}

func (c *Cron) GetTextract() {
	ctx := context.Background()
	utils.Log(ctx, "run cron get textract")

	jobs, err := repo.Find[model.Job](ctx, c.control.Db, constants.MODEL_JOB, primitive.M{
		"status": string(types.JobStatusInProgress),
	})
	if err != nil {
		utils.FilterError(ctx, err, "failed to get running job")
		return
	}

	blockChunks := map[int][]model.Block{}
	for _, job := range jobs {
		id, err := primitive.ObjectIDFromHex(job.DocumentId)
		if err != nil {
			utils.FilterError(ctx, err, "failed to parse document id")
			continue
		}
		mQuery := bson.M{"_id": id}

		doc := model.Document{}
		err = c.control.Db.FindOne(ctx, constants.MODEL_DOCUMENT, mQuery, &doc)
		if err != nil {
			utils.FilterError(ctx, err, "failed to get document by id")
			continue
		}

		if doc.ExtractType == constants.ExtractType_Text {
			allPages := false
			var nextToken *string

			page := 1
			blockChunkIdx := 1
			for !allPages {
				output, err := c.control.Textract.GetExtractText(ctx, job.JobId, nextToken)
				if err != nil {
					utils.FilterError(ctx, err, "failed to get extact text by job id")
					continue
				}

				if output.JobStatus == types.JobStatusSucceeded {
					blocks := helper.OutputBlockToModel(output.Blocks)

					if page%5 == 0 {
						blockChunkIdx++
					}
					blockChunks[blockChunkIdx] = append(blockChunks[blockChunkIdx], blocks...)

					if output.DocumentMetadata != nil {
						doc.DocumentMetadata = &model.DocumentMetadata{
							Pages: output.DocumentMetadata.Pages,
						}
					}
				}
				nextToken = output.NextToken
				job.Status = string(output.JobStatus)
				allPages = output.NextToken == nil

				page++
			}
		} else if doc.ExtractType == constants.ExtractType_Form {
			allPages := false
			var nextToken *string

			page := 1
			blockChunkIdx := 1
			for !allPages {
				output, err := c.control.Textract.GetExtractFormAndTable(ctx, job.JobId, nextToken)
				if err != nil {
					utils.FilterError(ctx, err, "failed to get extact text by job id")
					continue
				}

				if output.JobStatus == types.JobStatusSucceeded {
					utils.Log(ctx, fmt.Sprintf("extracting document %s page %d ", job.DocumentId, page))

					blocks := helper.OutputBlockToModel(output.Blocks)

					if page%5 == 0 {
						blockChunkIdx++
					}
					blockChunks[blockChunkIdx] = append(blockChunks[blockChunkIdx], blocks...)

					if output.DocumentMetadata != nil {
						doc.DocumentMetadata = &model.DocumentMetadata{
							Pages: output.DocumentMetadata.Pages,
						}
					}
				}
				nextToken = output.NextToken
				job.Status = string(output.JobStatus)
				allPages = output.NextToken == nil

				page++
			}
		} else {
			utils.Log(ctx, "invalid document type")
			continue
		}

		// kvSets := helper.MapFormType(doc.Blocks)
		// doc.MappedKeyValue = kvSets

		// tbSets := helper.MapTableType(doc.Blocks)
		// doc.MappedTables = tbSets

		docuemntStored := true
		if len(blockChunks) > 0 {
			for idx, chunk := range blockChunks {
				if idx == 1 {
					err := c.control.Db.UpdateOne(ctx, constants.MODEL_DOCUMENT, mQuery, bson.D{{Key: "$set", Value: bson.D{
						{Key: "blocks", Value: chunk},
						{Key: "documentMetadata", Value: doc.DocumentMetadata},
						// {Key: "mappedKeyValues", Value: doc.MappedKeyValue},
						// {Key: "mappedTables", Value: doc.MappedTables},
					}}})
					if err != nil {
						docuemntStored = false
						utils.FilterError(ctx, err, "failed to set document block")
						continue
					}
				} else {
					err := c.control.Db.InsertOne(ctx, constants.MODEL_DOCUMENT, model.Document{
						ID:               primitive.NewObjectID(),
						Name:             fmt.Sprintf("%s page %d", doc.Name, idx),
						Blocks:           chunk,
						DocumentMetadata: doc.DocumentMetadata,
						ExtractType:      doc.ExtractType,
					})
					if err != nil {
						docuemntStored = false
						utils.FilterError(ctx, err, "failed to set document block")
						continue
					}
				}
			}
		}
		if docuemntStored && job.Status != string(types.JobStatusInProgress) {
			err = c.control.Db.UpdateOne(ctx, constants.MODEL_JOB, bson.M{"_id": job.ID}, bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: job.Status}}}})
			if err != nil {
				utils.FilterError(ctx, err, "failed to set job status")
				continue
			}
		}
	}
}
