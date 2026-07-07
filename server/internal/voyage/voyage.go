package voyage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	apiKey string
	http   *http.Client
}

type EmbeddingRequest struct {
	Input      []string        `json:"input"`
	Model      string          `json:"model"`
	Input_Type VoyageInputType `json:"input_type"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type RerankRequest struct {
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	Model     string   `json:"model"`
	TopK      int      `json:"top_k,omitempty"`
}

type RerankResponse struct {
	Data []RerankResult `json:"data"`
}

type RerankResult struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type VoyageInputType string

const (
	Document VoyageInputType = "document"
	Query    VoyageInputType = "query"
)

func (i VoyageInputType) Valid() bool {
	switch i {
	case Document, Query:
		return true
	default:
		return false
	}
}

func New(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http:   &http.Client{},
	}
}

func (c *Client) Embed(ctx context.Context, inputType VoyageInputType, texts []string) ([][]float32, error) {

	if !inputType.Valid() {
		return nil, fmt.Errorf("invalid input type, please specify document or query")
	}

	voyageEmbeddingURL := "https://api.voyageai.com/v1/embeddings"

	payload := EmbeddingRequest{
		Input:      texts,
		Model:      "voyage-3",
		Input_Type: inputType,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("could not parse JSON into bytes: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, voyageEmbeddingURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create a request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed HTTP request %d: %v", res.StatusCode, body)
	}

	var out EmbeddingResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("had trouble unpacking the json: %w", err)
	}

	var embeddings [][]float32

	for _, val := range out.Data {
		embeddings = append(embeddings, val.Embedding)
	}

	return embeddings, nil
}

func (c *Client) Rerank(ctx context.Context, query string, docs []string, topK int) ([]RerankResult, error) {

	voyageRerankURL := "https://api.voyageai.com/v1/rerank"

	payload := RerankRequest{
		Query:     query,
		Documents: docs,
		Model:     "rerank-2.5",
		TopK:      topK,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert payload JSON %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, voyageRerankURL, bytes.NewReader(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("failed HTTP request: %d: %s", res.StatusCode, body)
	}

	var out RerankResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("couldn't turn res.Body into internal Rerank res: %w", err)
	}

	return out.Data, nil
}
