package analysis

import (
	"context"
	"fmt"
	"strings"

	"github.com/chessgoddess/chesslens/internal/engine"
	"github.com/chessgoddess/chesslens/internal/models"
	"github.com/chessgoddess/chesslens/internal/repository"
)

type Service struct {
	engine     *engine.Engine
	moveRepo   *repository.MoveRepository
	sessionRepo *repository.AnalysisSessionRepository
}

func NewService(stockfishPath string, moveRepo *repository.MoveRepository, sessionRepo *repository.AnalysisSessionRepository) (*Service, error) {
	eng, err := engine.New(stockfishPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize stockfish: %w", err)
	}
	
	if err := eng.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize engine: %w", err)
	}
	
	return &Service{
		engine:      eng,
		moveRepo:    moveRepo,
		sessionRepo: sessionRepo,
	}, nil
}

func (s *Service) AnalyzeGame(ctx context.Context, session *models.AnalysisSession, pgn string) error {
	if err := s.sessionRepo.UpdateStatus(ctx, session.ID, "running"); err != nil {
		return err
	}
	
	moves := extractMovesFromPGN(pgn)
	
	currentFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	
	for i, move := range moves {
		if move == "1-0" || move == "0-1" || move == "1/2-1/2" || move == "*" {
			continue
		}
		
		eval, err := s.engine.AnalyzeMove(currentFEN, move, session.Depth)
		if err != nil {
			continue
		}
		
		classification := classifyMove(eval.Eval)
		
		moveModel := &models.Move{
			SessionID:      session.ID,
			MoveNumber:     i + 1,
			FEN:            currentFEN,
			SAN:            move,
			Evaluation:     eval.Eval,
			BestMove:       eval.BestMove,
			Classification: classification,
			Depth:          eval.Depth,
		}
		
		if err := s.moveRepo.Create(ctx, moveModel); err != nil {
			return err
		}
		
		currentFEN = applyMove(currentFEN, move)
	}
	
	return s.sessionRepo.UpdateStatus(ctx, session.ID, "completed")
}

func extractMovesFromPGN(pgn string) []string {
	lines := strings.Split(pgn, "\n")
	var moveText string
	inMoves := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" && inMoves {
			continue
		}
		if strings.HasPrefix(line, "[") {
			continue
		}
		inMoves = true
		moveText += " " + line
	}
	
	var moves []string
	parts := strings.Fields(moveText)
	
	for _, part := range parts {
		if strings.Contains(part, ".") {
			parts := strings.Split(part, ".")
			if len(parts) == 2 && parts[1] != "" {
				moves = append(moves, parts[1])
			}
		} else if part != "" && part != "1-0" && part != "0-1" && part != "1/2-1/2" && part != "*" {
			moves = append(moves, part)
		} else if part == "1-0" || part == "0-1" || part == "1/2-1/2" || part == "*" {
			moves = append(moves, part)
		}
	}
	
	return moves
}

func classifyMove(eval float64) string {
	absEval := eval
	if absEval < 0 {
		absEval = -absEval
	}
	
	switch {
	case absEval > 3.0:
		return "blunder"
	case absEval > 1.5:
		return "mistake"
	case absEval > 0.5:
		return "inaccuracy"
	case absEval > 0.2:
		return "good"
	default:
		return "best"
	}
}

func applyMove(fen, move string) string {
	// TODO: Implement proper FEN update using chess.js logic
	// For now, return the same FEN (placeholder)
	return fen
}

func (s *Service) Close() error {
	return s.engine.Close()
}
