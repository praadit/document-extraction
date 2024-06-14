package request

import "mime/multipart"

type AnalyzeDocumentRequest struct {
	File         *multipart.FileHeader `form:"file"`
	ExtractType  string                `form:"extractType" binding:"required"`
	S3ObjectName string                `form:"name"`
}
type ProcessResultRequest struct {
	Key string `json:"key"`
	Id  string `json:"id"`
}

type AskRequest struct {
	Query string `json:"query"`
}
