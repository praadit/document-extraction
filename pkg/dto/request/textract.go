package request

import "mime/multipart"

type AnalyzeDocumentRequest struct {
	File        *multipart.FileHeader `form:"file" binding:"required"`
	ExtractType string                `form:"extractType" binding:"required"`
}
type ProcessResultRequest struct {
	Key string `json:"key"`
	Id  string `json:"id"`
}
