package ai

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaChunk struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Thinking string `json:"thinking"`
	Done     bool   `json:"done"`
}

type EmbedBatchRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbedBatchResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}
