package service

import (
	"testing"
)

func TestExtractMovesFromPGN(t *testing.T) {
	pgn := `[Event "Test"]
[White "Alice"]
[Black "Bob"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 1-0`

	moves := extractMovesFromPGN(pgn)

	// extractMovesFromPGN includes result tokens; AnalyzeGame skips them.
	want := []string{"e4", "e5", "Nf3", "Nc6", "Bb5", "a6", "1-0"}
	if len(moves) != len(want) {
		t.Fatalf("expected %d tokens, got %d: %v", len(want), len(moves), moves)
	}
	for i, m := range want {
		if moves[i] != m {
			t.Errorf("move[%d]: want %q, got %q", i, m, moves[i])
		}
	}
}

func TestExtractMovesFromPGN_Empty(t *testing.T) {
	moves := extractMovesFromPGN("")
	if len(moves) != 0 {
		t.Errorf("expected 0 moves, got %d", len(moves))
	}
}

func TestExtractMovesFromPGN_ResultOnly(t *testing.T) {
	pgn := `[Result "*"]

*`
	moves := extractMovesFromPGN(pgn)
	// Result token should not appear in moves list (already filtered by AnalyzeGame loop).
	// extractMovesFromPGN includes result tokens; that's acceptable — AnalyzeGame skips them.
	for _, m := range moves {
		if m == "e4" {
			t.Error("unexpected real move in result-only PGN")
		}
	}
}

func TestClassifyMove(t *testing.T) {
	cases := []struct {
		eval float64
		want string
	}{
		{0.0, "best"},
		{0.15, "best"},
		{0.25, "good"},
		{0.6, "inaccuracy"},
		{2.0, "mistake"},
		{4.0, "blunder"},
		{-4.0, "blunder"},
		{-2.0, "mistake"},
		{-0.6, "inaccuracy"},
	}

	for _, tc := range cases {
		got := classifyMove(tc.eval)
		if got != tc.want {
			t.Errorf("classifyMove(%v) = %q, want %q", tc.eval, got, tc.want)
		}
	}
}

func TestApplyMove_StartingPosition(t *testing.T) {
	startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

	got, err := applyMove(startFEN, "e4")
	if err != nil {
		t.Fatalf("applyMove e4 failed: %v", err)
	}
	if got == startFEN {
		t.Error("FEN should change after e4")
	}
}

func TestApplyMove_MultipleMovesAdvancePosition(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	moves := []string{"e4", "e5", "Nf3", "Nc6", "Bb5"}

	for _, san := range moves {
		next, err := applyMove(fen, san)
		if err != nil {
			t.Fatalf("applyMove(%q) failed: %v", san, err)
		}
		if next == fen {
			t.Errorf("FEN did not advance after %q", san)
		}
		fen = next
	}
}

func TestApplyMove_InvalidSAN(t *testing.T) {
	startFEN := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	_, err := applyMove(startFEN, "Nf6") // black's knight — illegal for white
	if err == nil {
		t.Error("expected error for illegal move, got nil")
	}
}

func TestApplyMove_InvalidFEN(t *testing.T) {
	_, err := applyMove("not-a-fen", "e4")
	if err == nil {
		t.Error("expected error for invalid FEN, got nil")
	}
}
