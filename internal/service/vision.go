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
	"strings"
	"time"

	"github.com/notnil/chess"
)

var defaultVisionModels = []string{
	"google/gemma-4-31b-it:free",
	"google/gemma-4-26b-a4b-it:free",
	"nvidia/nemotron-nano-12b-v2-vl:free",
	"nvidia/nemotron-3-nano-omni-30b-a3b-reasoning:free",
}

// allowedImageMIME lists the image formats accepted for board recognition.
var allowedImageMIME = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/webp": true,
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

func (c *VisionClient) imageToFENWithModel(ctx context.Context, model, dataURL string) (string, int, error) {
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
						"url": dataURL,
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
	mime, err := detectImageMIME(imageData)
	if err != nil {
		return "", err
	}
	dataURL := fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(imageData))

	for i, model := range c.models {
		result, statusCode, err := c.imageToFENWithModel(ctx, model, dataURL)
		if err != nil {
			if statusCode == http.StatusTooManyRequests || statusCode >= 500 {
				slog.Warn("vision model unavailable, trying next", "model", model, "status", statusCode)
				continue
			}
			return "", err
		}

		fen := normalizeFEN(strings.TrimSpace(result))
		if !isValidFEN(fen) {
			slog.Warn("vision model returned invalid FEN, trying next", "model", model, "raw", result)
			continue
		}
		if i > 0 {
			slog.Info("vision fallback succeeded", "model", model, "attempt", i+1)
		}
		return fen, nil
	}
	return "", fmt.Errorf("all %d vision models exhausted or returned an unreadable position", len(c.models))
}

// detectImageMIME sniffs the image format from its bytes and rejects
// anything that is not a supported chess-board image.
func detectImageMIME(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty image data")
	}
	mime := http.DetectContentType(data)
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	if !allowedImageMIME[mime] {
		return "", fmt.Errorf("unsupported image format %q — use PNG, JPEG, or WebP", mime)
	}
	return mime, nil
}

// normalizeFEN pads a possibly-partial FEN to the full six fields,
// filling sensible defaults — vision models often omit trailing fields.
func normalizeFEN(fen string) string {
	fields := strings.Fields(fen)
	if len(fields) == 0 {
		return fen
	}
	defaults := []string{"", "w", "KQkq", "-", "0", "1"}
	for len(fields) < 6 {
		fields = append(fields, defaults[len(fields)])
	}
	return strings.Join(fields[:6], " ")
}

// isValidFEN reports whether fen parses as a legal chess position.
func isValidFEN(fen string) bool {
	if fen == "" {
		return false
	}
	_, err := chess.FEN(fen)
	return err == nil
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
