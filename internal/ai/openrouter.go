package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OpenRouterClient struct {
	apiKey    string
	baseURL   string
	model     string
	httpClient *http.Client
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func NewOpenRouterClient(apiKey string, model string) *OpenRouterClient {
	if model == "" {
		model = "openai/gpt-4o-mini"
	}

	return &OpenRouterClient{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *OpenRouterClient) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   1000,
		Temperature: 0.7,
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

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *OpenRouterClient) ExplainMove(fen, move, classification string, eval float64) (string, error) {
	prompt := buildMoveExplanationPrompt(fen, move, classification, eval)
	
	messages := []ChatMessage{
		{
			Role:    "system",
			Content: "You are a chess coach explaining moves to a student. Be concise, insightful, and focus on key positional/tactical ideas. Keep explanations under 100 words.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	return c.Chat(context.Background(), messages)
}

func (c *OpenRouterClient) ExplainBlunder(fen, move, bestMove string, evalBefore, evalAfter float64) (string, error) {
	prompt := fmt.Sprintf(
		"I played %s in this position: %s\n\nThe evaluation dropped from %.2f to %.2f.\nThe best move was %s.\n\nWhy is my move a blunder? What did I miss? Explain in simple terms.",
		move, fen, evalBefore, evalAfter, bestMove,
	)

	messages := []ChatMessage{
		{
			Role:    "system",
			Content: "You are a chess coach explaining why a move is a blunder. Be specific about tactical/positional consequences. Keep it under 150 words.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	return c.Chat(context.Background(), messages)
}

func (c *OpenRouterClient) SummarizeGame(opening, result string, blunderCount, mistakeCount int) (string, error) {
	prompt := fmt.Sprintf(
		"Opening: %s\nResult: %s\nBlunders: %d\nMistakes: %d\n\nGive a 2-3 sentence summary of this game's key moments and learning points.",
		opening, result, blunderCount, mistakeCount,
	)

	messages := []ChatMessage{
		{
			Role:    "system",
			Content: "You are a chess coach summarizing a game. Focus on key turning points and learning opportunities. Keep it under 100 words.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	return c.Chat(context.Background(), messages)
}

func buildMoveExplanationPrompt(fen, move, classification string, eval float64) string {
	classificationContext := ""
	switch classification {
	case "blunder":
		classificationContext = "This move is a blunder (major mistake)."
	case "mistake":
		classificationContext = "This move is a mistake."
	case "inaccuracy":
		classificationContext = "This move is slightly inaccurate."
	case "best":
		classificationContext = "This is the best move according to the engine."
	case "excellent":
		classificationContext = "This is an excellent move."
	}

	return fmt.Sprintf(
		"Position (FEN): %s\nMove played: %s\nEvaluation: %.2f\n%s\n\nExplain the reasoning behind this move and what the player should consider.",
		fen, move, eval, classificationContext,
	)
}
