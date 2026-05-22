// Package service contains business logic for game analysis, AI explanations, and vision.
package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/chessgoddess/chessgoddess/internal/engine"
	"github.com/chessgoddess/chessgoddess/internal/model"
	"github.com/chessgoddess/chessgoddess/internal/repository"
	"github.com/notnil/chess"
)

const (
	startFEN        = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	analysisMultiPV = 3
)

type AnalysisService struct {
	pool        *engine.Pool
	moveRepo    *repository.MoveRepository
	sessionRepo *repository.AnalysisSessionRepository
}

func NewAnalysisService(stockfishPath string, poolSize int, moveRepo *repository.MoveRepository, sessionRepo *repository.AnalysisSessionRepository) (*AnalysisService, error) {
	pool, err := engine.NewPool(stockfishPath, poolSize)
	if err != nil {
		return nil, fmt.Errorf("failed to start engine pool: %w", err)
	}
	return &AnalysisService{
		pool:        pool,
		moveRepo:    moveRepo,
		sessionRepo: sessionRepo,
	}, nil
}

// Pool exposes the engine pool for on-demand position evaluation.
func (s *AnalysisService) Pool() *engine.Pool { return s.pool }

// AnalyzeGame replays a PGN, evaluates every position, and stores per-move
// metrics (centipawn loss, classification, accuracy) plus per-side accuracy.
func (s *AnalysisService) AnalyzeGame(ctx context.Context, session *model.AnalysisSession, pgn string) error {
	if err := s.sessionRepo.UpdateStatus(ctx, session.ID, "running"); err != nil {
		return err
	}

	// Replay the PGN into a list of FENs; stop at the first illegal move.
	sans := playableMoves(extractMovesFromPGN(pgn))
	fens := []string{startFEN}
	current := startFEN
	valid := make([]string, 0, len(sans))
	for _, san := range sans {
		next, err := applyMove(current, san)
		if err != nil {
			break
		}
		valid = append(valid, san)
		fens = append(fens, next)
		current = next
	}
	sans = valid

	if len(sans) == 0 {
		return s.failSession(ctx, session.ID, fmt.Errorf("no legal moves found in PGN"))
	}

	// Evaluate every position concurrently — the engine pool bounds concurrency.
	evals := make([]*engine.Evaluation, len(fens))
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var firstErr error
	for i, fen := range fens {
		wg.Add(1)
		go func(i int, fen string) {
			defer wg.Done()
			if term := terminalEval(fen); term != nil {
				evals[i] = term
				return
			}
			ev, err := s.pool.Evaluate(ctx, fen, session.Depth, analysisMultiPV)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errMu.Unlock()
				return
			}
			evals[i] = ev
		}(i, fen)
	}
	wg.Wait()
	if firstErr != nil {
		return s.failSession(ctx, session.ID, firstErr)
	}

	var whiteAcc, blackAcc []float64

	for i := 1; i <= len(sans); i++ {
		before := evals[i-1] // side to move = the player who made move i
		after := evals[i]    // side to move = the opponent
		moverIsWhite := i%2 == 1

		bestScore := before.Lines[0].Score()   // mover POV, centipawns
		playedScore := -after.Lines[0].Score() // mover POV (negate opponent POV)
		cpLoss := float64(bestScore - playedScore)
		if cpLoss < 0 {
			cpLoss = 0
		}

		bestUCI := ""
		if len(before.Lines[0].PV) > 0 {
			bestUCI = before.Lines[0].PV[0]
		}
		playedUCI, _ := sanToUCI(fens[i-1], sans[i-1])
		isBest := bestUCI != "" && playedUCI == bestUCI

		class := classify(cpLoss, i, isBest, isOnlyGoodMove(before), bestScore)

		// Win-probability accuracy from the mover's POV.
		acc := moveAccuracy(winPercent(float64(bestScore)), winPercent(float64(playedScore)))
		if moverIsWhite {
			whiteAcc = append(whiteAcc, acc)
		} else {
			blackAcc = append(blackAcc, acc)
		}

		evalBeforeWhite := toWhitePOV(bestScore, moverIsWhite)
		evalAfterWhite := toWhitePOV(after.Lines[0].Score(), !moverIsWhite)

		m := &model.Move{
			SessionID:      session.ID,
			MoveNumber:     i,
			FEN:            fens[i-1],
			SAN:            sans[i-1],
			Evaluation:     evalAfterWhite / 100.0,
			EvalBefore:     evalBeforeWhite / 100.0,
			EvalAfter:      evalAfterWhite / 100.0,
			CPLoss:         cpLoss,
			Accuracy:       acc,
			BestMove:       bestUCI,
			BestLine:       strings.Join(before.Lines[0].PV, " "),
			Classification: class,
			Depth:          before.Depth,
		}
		if err := s.moveRepo.Create(ctx, m); err != nil {
			return s.failSession(ctx, session.ID, err)
		}
	}

	if err := s.sessionRepo.UpdateAccuracy(ctx, session.ID, mean(whiteAcc), mean(blackAcc)); err != nil {
		return err
	}
	return s.sessionRepo.UpdateStatus(ctx, session.ID, "completed")
}

func (s *AnalysisService) GetMovesBySessionID(ctx context.Context, sessionID string) ([]model.Move, error) {
	return s.moveRepo.ListBySessionID(ctx, sessionID)
}

func (s *AnalysisService) Close() error {
	return s.pool.Close()
}

func (s *AnalysisService) failSession(ctx context.Context, id string, cause error) error {
	_ = s.sessionRepo.UpdateStatus(ctx, id, "failed")
	return cause
}

// classify assigns a chess.com-style label from the centipawn loss of a move.
func classify(cpLoss float64, ply int, isBest, onlyGood bool, evalBeforeMover int) string {
	switch {
	case isBest && onlyGood && evalBeforeMover <= 50:
		return "brilliant" // the only good move, found in a non-winning position
	case isBest && onlyGood:
		return "great" // the only move that holds the advantage
	case isBest:
		return "best"
	case ply <= 16 && cpLoss <= 30:
		return "book" // quiet, sound opening move (no opening DB — heuristic)
	case cpLoss <= 10:
		return "excellent"
	case cpLoss <= 40:
		return "good"
	case cpLoss <= 90:
		return "inaccuracy"
	case cpLoss <= 180:
		return "mistake"
	default:
		return "blunder"
	}
}

// isOnlyGoodMove reports whether the best line is decisively better than every
// alternative — i.e. the position has a single move that holds.
func isOnlyGoodMove(ev *engine.Evaluation) bool {
	if ev == nil || len(ev.Lines) < 2 {
		return false
	}
	best := ev.Lines[0].Score()
	second := ev.Lines[1].Score()
	return best-second >= 150 && second <= 30
}

// winPercent maps a centipawn score to a 0-100 win probability for that side.
func winPercent(cp float64) float64 {
	return 50 + 50*(2/(1+math.Exp(-0.00368208*cp))-1)
}

// moveAccuracy converts the win-probability drop a move caused into a 0-100 score.
func moveAccuracy(winBefore, winAfter float64) float64 {
	loss := winBefore - winAfter
	if loss < 0 {
		loss = 0
	}
	acc := 103.1668*math.Exp(-0.04354*loss) - 3.1669
	switch {
	case acc > 100:
		return 100
	case acc < 0:
		return 0
	default:
		return acc
	}
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}

// toWhitePOV converts a side-to-move centipawn score to White's POV.
func toWhitePOV(scoreSTM int, whiteToMove bool) float64 {
	if whiteToMove {
		return float64(scoreSTM)
	}
	return float64(-scoreSTM)
}

// playableMoves drops PGN result tokens, keeping only real moves.
func playableMoves(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		switch t {
		case "1-0", "0-1", "1/2-1/2", "*":
			continue
		default:
			out = append(out, t)
		}
	}
	return out
}

// terminalEval returns a synthetic evaluation for a game-over position,
// or nil if the position is still playable.
func terminalEval(fen string) *engine.Evaluation {
	fenOpt, err := chess.FEN(fen)
	if err != nil {
		return nil
	}
	g := chess.NewGame(fenOpt)
	switch g.Outcome() {
	case chess.WhiteWon, chess.BlackWon:
		// The side to move has been checkmated.
		return &engine.Evaluation{FEN: fen, Lines: []engine.Line{{Rank: 1, ScoreCP: -engine.MateScore}}}
	case chess.Draw:
		return &engine.Evaluation{FEN: fen, Lines: []engine.Line{{Rank: 1, ScoreCP: 0}}}
	default:
		return nil
	}
}

// sanToUCI converts a SAN move played from fen into UCI long-algebraic form.
func sanToUCI(fen, san string) (string, error) {
	fenOpt, err := chess.FEN(fen)
	if err != nil {
		return "", err
	}
	g := chess.NewGame(fenOpt)
	if err := g.MoveStr(san); err != nil {
		return "", err
	}
	moves := g.Moves()
	if len(moves) == 0 {
		return "", fmt.Errorf("no move recorded for %q", san)
	}
	return moves[len(moves)-1].String(), nil
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
		} else if part != "" {
			moves = append(moves, part)
		}
	}

	return moves
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
