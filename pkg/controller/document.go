package controller

import (
	"fmt"
	"net/http"
	"path/filepath"
	constants "textract-mongo/pkg/const"
	"textract-mongo/pkg/controller/helper"
	"textract-mongo/pkg/dto/request"
	"textract-mongo/pkg/dto/response"
	"textract-mongo/pkg/model"
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
	awsS3name := "textract/" + filename

	if err := c.s3.Upload(ctx, utils.Config.AwsS3Bucket, awsS3name, file); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
			Message: utils.FilterError(ctx, err, "failed to upload file").Error(),
		})
		return
	}

	doc := model.Document{
		ID:          primitive.NewObjectID(),
		Name:        filename,
		ExtarctType: body.ExtractType,
	}
	if body.ExtractType == constants.ExtractType_Text {
		output, err := c.textract.ExtractText(ctx, nil, &utils.Config.AwsS3Bucket, &awsS3name)
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
		output, err := c.textract.ExtractFormAndTable(ctx, nil, &utils.Config.AwsS3Bucket, &awsS3name)
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

	err = c.db.InsertOne(ctx, constants.MODEL_DOCUMENT, doc)
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
	err := c.db.FindOne(ctx, constants.MODEL_DOCUMENT, mQuery, &doc)
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
	if doc.ExtarctType == constants.ExtractType_Form {
		sum, err := c.ollama.SummarizeForm(ctx, textToSummarize, doc.MappedKeyValue)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to summarize form document").Error(),
			})
			return
		}

		summary = sum
	} else if doc.ExtarctType == constants.ExtractType_Text {
		sum, err := c.ollama.SummarizeText(ctx, textToSummarize)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.BaseResponse{
				Message: utils.FilterError(ctx, err, "failed to summarize form document").Error(),
			})
			return
		}

		summary = sum
	}

	err = c.db.UpdateOne(ctx, constants.MODEL_DOCUMENT, mQuery, bson.D{{Key: "$set", Value: bson.D{{Key: "summary", Value: summary}}}})
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