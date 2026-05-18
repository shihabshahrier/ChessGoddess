package vision

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ImageToFENClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

type VisionMessage struct {
	Role    string `json:"role"`
	Content []any  `json:"content"`
}

type VisionRequest struct {
	Model    string          `json:"model"`
	Messages []VisionMessage `json:"messages"`
	MaxTokens int            `json:"max_tokens"`
}

type VisionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewImageToFENClient(apiKey string, model string) *ImageToFENClient {
	if model == "" {
		model = "openai/gpt-4o"
	}

	return &ImageToFENClient{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *ImageToFENClient) ImageToFEN(ctx context.Context, imageData []byte) (string, error) {
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
		Model:     c.model,
		Messages:  messages,
		MaxTokens: 100,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("HTTP-Referer", "https://chesslens.app")
	httpReq.Header.Set("X-Title", "ChessLens")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var visionResp VisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&visionResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(visionResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	fen := visionResp.Choices[0].Message.Content
	fen = extractFEN(fen)

	return fen, nil
}

func extractFEN(response string) string {
	response = response
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
