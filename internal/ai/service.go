package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/chessgoddess/chesslens/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	client    *OpenRouterClient
	cache     *redis.Client
	repo      *repository.AIExplanationRepository
	ctx       context.Context
}

func NewService(apiKey string, model string, redisClient *redis.Client, repo *repository.AIExplanationRepository) *Service {
	return &Service{
		client: NewOpenRouterClient(apiKey, model),
		cache:  redisClient,
		repo:   repo,
		ctx:    context.Background(),
	}
}

func (s *Service) getCacheKey(fen, move, promptType string) string {
	hash := sha256.Sum256([]byte(fen + move + promptType))
	return "ai:cache:" + hex.EncodeToString(hash[:])
}

func (s *Service) getCached(fen, move, promptType string) (string, error) {
	if s.cache == nil {
		return "", fmt.Errorf("cache not available")
	}

	key := s.getCacheKey(fen, move, promptType)
	val, err := s.cache.Get(s.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func (s *Service) setCache(fen, move, promptType, content string) error {
	if s.cache == nil {
		return nil
	}

	key := s.getCacheKey(fen, move, promptType)
	return s.cache.Set(s.ctx, key, content, 0).Err()
}

func (s *Service) ExplainMove(ctx context.Context, sessionID, moveID, fen, move, classification string, eval float64) (string, error) {
	cached, err := s.getCached(fen, move, "explain")
	if err == nil && cached != "" {
		return cached, nil
	}

	explanation, err := s.client.ExplainMove(fen, move, classification, eval)
	if err != nil {
		return "", err
	}

	s.repo.Create(ctx, sessionID, moveID, fen, explanation, s.client.model)
	s.setCache(fen, move, "explain", explanation)

	return explanation, nil
}

func (s *Service) ExplainBlunder(ctx context.Context, sessionID, moveID, fen, move, bestMove string, evalBefore, evalAfter float64) (string, error) {
	cached, err := s.getCached(fen, move, "blunder")
	if err == nil && cached != "" {
		return cached, nil
	}

	explanation, err := s.client.ExplainBlunder(fen, move, bestMove, evalBefore, evalAfter)
	if err != nil {
		return "", err
	}

	s.repo.Create(ctx, sessionID, moveID, fen, explanation, s.client.model)
	s.setCache(fen, move, "blunder", explanation)

	return explanation, nil
}

func (s *Service) SummarizeGame(ctx context.Context, sessionID, opening, result string, blunderCount, mistakeCount int) (string, error) {
	cached, err := s.getCached(opening, result, "summary")
	if err == nil && cached != "" {
		return cached, nil
	}

	summary, err := s.client.SummarizeGame(opening, result, blunderCount, mistakeCount)
	if err != nil {
		return "", err
	}

	s.repo.Create(ctx, sessionID, "", opening, summary, s.client.model)
	s.setCache(opening, result, "summary", summary)

	return summary, nil
}

func (s *Service) GetRepo() *repository.AIExplanationRepository {
	return s.repo
}
