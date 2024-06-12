package response

type GetTextResponse struct {
	ExtractedText any `json:"extractedText"`
}
type AnalyzeResponse struct {
	Name   string `json:"name"`
	Blocks any    `json:"blokcs"`
}
type SummarizedResponse struct {
	SummarizedText any `json:"summarize"`
}
type ClassifyResponse struct {
	Classification any `json:"classification"`
}
