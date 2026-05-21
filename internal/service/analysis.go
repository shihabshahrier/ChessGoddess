// Package service contains business logic for game analysis, AI explanations, and vision.
package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/chessgoddess/chessgoddess/internal/engine"
	"github.com/chessgoddess/chessgoddess/internal/model"
	"github.com/chessgoddess/chessgoddess/internal/repository"
	"github.com/notnil/chess"
)

type AnalysisService struct {
	engine      *engine.Engine
	moveRepo    *repository.MoveRepository
	sessionRepo *repository.AnalysisSessionRepository
}

func NewAnalysisService(stockfishPath string, moveRepo *repository.MoveRepository, sessionRepo *repository.AnalysisSessionRepository) (*AnalysisService, error) {
	eng, err := engine.New(stockfishPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize stockfish: %w", err)
	}

	if err := eng.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize engine: %w", err)
	}

	return &AnalysisService{
		engine:      eng,
		moveRepo:    moveRepo,
		sessionRepo: sessionRepo,
	}, nil
}

func (s *AnalysisService) AnalyzeGame(ctx context.Context, session *model.AnalysisSession, pgn string) error {
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

		moveModel := &model.Move{
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

		nextFEN, err := applyMove(currentFEN, move)
		if err != nil {
			// Invalid SAN — skip position advance but continue analysis.
			continue
		}
		currentFEN = nextFEN
	}

	return s.sessionRepo.UpdateStatus(ctx, session.ID, "completed")
}

func (s *AnalysisService) GetMovesBySessionID(ctx context.Context, sessionID string) ([]model.Move, error) {
	return s.moveRepo.ListBySessionID(ctx, sessionID)
}

func (s *AnalysisService) Close() error {
	return s.engine.Close()
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
			segments := strings.Split(part, ".")
			if len(segments) == 2 && segments[1] != "" {
				moves = append(moves, segments[1])
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

// applyMove advances position by one SAN move, returning the resulting FEN.
func applyMove(fen, san string) (string, error) {
	fenOpt, err := chess.FEN(fen)
	if err != nil {
		return fen, fmt.Errorf("invalid FEN: %w", err)
	}
	game := chess.NewGame(fenOpt)
	if err := game.MoveStr(san); err != nil {
		return fen, fmt.Errorf("illegal move %q: %w", san, err)
	}
	return game.Position().String(), nil
}
