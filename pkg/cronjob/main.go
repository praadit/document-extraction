package cronjob

import (
	"context"
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

	jobs, err := repo.Find[model.Job](ctx, c.control.Db, constants.MODEL_JOB, primitive.M{
		"status": string(types.JobStatusInProgress),
	})
	if err != nil {
		utils.FilterError(ctx, err, "failed to get running job")
		return
	}

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
			output, err := c.control.Textract.GetExtractText(ctx, job.JobId)
			if err != nil {
				utils.FilterError(ctx, err, "failed to get extact text by job id")
				continue
			}

			if output.JobStatus == types.JobStatusSucceeded {
				blocks := helper.OutputBlockToModel(output.Blocks)
				doc.Blocks = blocks
				if output.DocumentMetadata != nil {
					doc.DocumentMetadata = &model.DocumentMetadata{
						Pages: output.DocumentMetadata.Pages,
					}
				}
				job.Status = string(output.JobStatus)
			} else if output.JobStatus == types.JobStatusInProgress {
				job.Status = string(output.JobStatus)
			}
		} else if doc.ExtractType == constants.ExtractType_Form {
			output, err := c.control.Textract.GetExtractFormAndTable(ctx, job.JobId)
			if err != nil {
				utils.FilterError(ctx, err, "failed to get extact form & table by job id")
				continue
			}

			if output.JobStatus == types.JobStatusSucceeded {
				blocks := helper.OutputBlockToModel(output.Blocks)
				doc.Blocks = blocks
				if output.DocumentMetadata != nil {
					doc.DocumentMetadata = &model.DocumentMetadata{
						Pages: output.DocumentMetadata.Pages,
					}
				}
				job.Status = string(output.JobStatus)
			} else if output.JobStatus == types.JobStatusInProgress {
				job.Status = string(output.JobStatus)
			}
		} else {
			utils.Log(ctx, "invalid document type")
			continue
		}

		if len(doc.Blocks) > 0 {
			err = c.control.Db.UpdateOne(ctx, constants.MODEL_DOCUMENT, mQuery, bson.D{{Key: "$set", Value: bson.D{{Key: "blocks", Value: doc.Blocks}, {Key: "documentMetadata", Value: doc.DocumentMetadata}}}})
			if err != nil {
				utils.FilterError(ctx, err, "failed to set document block")
				continue
			}
		}
		if job.Status != string(types.JobStatusInProgress) {
			err = c.control.Db.UpdateOne(ctx, constants.MODEL_JOB, bson.M{"_id": job.ID}, bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: job.Status}}}})
			if err != nil {
				utils.FilterError(ctx, err, "failed to set job status")
				continue
			}
		}
	}
}
