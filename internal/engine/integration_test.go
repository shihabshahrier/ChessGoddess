package engine

import (
	"context"
	"os/exec"
	"testing"
)

func requireStockfish(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("stockfish")
	if err != nil {
		t.Skip("stockfish not installed — skipping engine integration test")
	}
	return path
}

func TestEngineEvaluateStartpos(t *testing.T) {
	path := requireStockfish(t)
	pool, err := NewPool(path, 2)
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()

	eval, err := pool.Evaluate(
		context.Background(),
		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		12, 3,
	)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if len(eval.Lines) == 0 {
		t.Fatal("expected at least one line")
	}
	if eval.BestMove == "" {
		t.Error("expected a best move")
	}
	if eval.Lines[0].Depth < 12 {
		t.Errorf("depth %d, want >= 12", eval.Lines[0].Depth)
	}
	if len(eval.Lines[0].PV) == 0 {
		t.Error("best line has no PV")
	}
}

func TestEngineEvaluateMate(t *testing.T) {
	path := requireStockfish(t)
	e, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer e.Close()
	if err := e.Initialize(); err != nil {
		t.Fatalf("Initialize: %v", err)
	}

	// White to move: Ra8 is checkmate.
	eval, err := e.Evaluate("7k/5ppp/8/8/8/8/8/R6K w - - 0 1", 14, 1)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if eval.Lines[0].Mate <= 0 {
		t.Errorf("expected a positive mate score, got mate=%d cp=%d",
			eval.Lines[0].Mate, eval.Lines[0].ScoreCP)
	}
}
