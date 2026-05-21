// Vision service — converts chess board images to FEN strings via OpenRouter with fallback.
package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

var defaultVisionModels = []string{
	"google/gemma-4-31b-it:free",
	"google/gemma-4-26b-a4b-it:free",
	"nvidia/nemotron-nano-12b-v2-vl:free",
	"nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free",
}

type VisionClient struct {
	apiKey     string
	baseURL    string
	models     []string
	httpClient *http.Client
}

type VisionMessage struct {
	Role    string `json:"role"`
	Content []any  `json:"content"`
}

type VisionRequest struct {
	Model     string          `json:"model"`
	Messages  []VisionMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

type VisionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewVisionClient(apiKey string, model string) *VisionClient {
	models := defaultVisionModels
	if model != "" {
		models = append([]string{model}, defaultVisionModels...)
	}
	return &VisionClient{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		models:  models,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *VisionClient) imageToFENWithModel(ctx context.Context, model string, imageData []byte) (string, int, error) {
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	messages := []VisionMessage{
		{
			Role: "system",
			Content: []any{
				map[string]string{
					"type": "text",
					"text": "You are a chess position analyzer. Look at the chess board image and return ONLY the FEN string representing the position. Do not include any explanation, just the FEN.",
				},
			},
		},
		{
			Role: "user",
			Content: []any{
				map[string]string{
					"type": "text",
					"text": "What is the FEN of this chess position?",
				},
				map[string]any{
					"type": "image_url",
					"image_url": map[string]string{
						"url": fmt.Sprintf("data:image/png;base64,%s", base64Image),
					},
				},
			},
		},
	}

	req := VisionRequest{
		Model:     model,
		Messages:  messages,
		MaxTokens: 100,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://chessgoddess.app")
	httpReq.Header.Set("X-Title", "ChessGoddess")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		return "", resp.StatusCode, fmt.Errorf("API error %d", resp.StatusCode)
	}

	var visionResp VisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&visionResp); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(visionResp.Choices) == 0 {
		return "", 0, fmt.Errorf("no choices in response")
	}

	return extractFEN(visionResp.Choices[0].Message.Content), http.StatusOK, nil
}

func (c *VisionClient) ImageToFEN(ctx context.Context, imageData []byte) (string, error) {
	for i, model := range c.models {
		result, statusCode, err := c.imageToFENWithModel(ctx, model, imageData)
		if err == nil {
			if i > 0 {
				slog.Info("vision fallback succeeded", "model", model, "attempt", i+1)
			}
			return result, nil
		}

		if statusCode == http.StatusTooManyRequests || statusCode >= 500 {
			slog.Warn("vision model unavailable, trying next", "model", model, "status", statusCode)
			continue
		}

		return "", err
	}
	return "", fmt.Errorf("all %d vision models exhausted", len(c.models))
}

func extractFEN(response string) string {
	for i := 0; i < 6; i++ {
		if idx := bytes.IndexByte([]byte(response), '/'); idx >= 0 {
			start := idx - 1
			for start > 0 && response[start-1] != ' ' && response[start-1] != '\n' {
				start--
			}

			end := start
			spaceCount := 0
			for end < len(response) && spaceCount < 5 {
				if response[end] == ' ' {
					spaceCount++
				}
				end++
			}

			if spaceCount >= 5 {
				return response[start:end]
			}
		}
	}
	return response
}
