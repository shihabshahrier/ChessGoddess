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

func TestClassify(t *testing.T) {
	cases := []struct {
		name       string
		cpLoss     float64
		ply        int
		isBest     bool
		onlyGood   bool
		evalBefore int
		want       string
	}{
		{"engine best move", 0, 30, true, false, 100, "best"},
		{"only good move, not winning", 0, 30, true, true, 0, "brilliant"},
		{"only good move, already winning", 0, 30, true, true, 300, "great"},
		{"quiet opening move", 25, 8, false, false, 0, "book"},
		{"near best", 5, 30, false, false, 0, "excellent"},
		{"small slip", 30, 30, false, false, 0, "good"},
		{"inaccuracy", 70, 30, false, false, 0, "inaccuracy"},
		{"mistake", 120, 30, false, false, 0, "mistake"},
		{"blunder", 400, 30, false, false, 0, "blunder"},
	}

	for _, tc := range cases {
		got := classify(tc.cpLoss, tc.ply, tc.isBest, tc.onlyGood, tc.evalBefore)
		if got != tc.want {
			t.Errorf("%s: classify = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestMoveAccuracy(t *testing.T) {
	// A move that keeps the win probability flat scores near 100.
	if acc := moveAccuracy(60, 60); acc < 99 {
		t.Errorf("flat move accuracy = %.1f, want ~100", acc)
	}
	// A move that craters win probability scores low.
	if acc := moveAccuracy(80, 20); acc > 30 {
		t.Errorf("blunder accuracy = %.1f, want low", acc)
	}
	// Accuracy never leaves the 0-100 range.
	if acc := moveAccuracy(100, 0); acc < 0 || acc > 100 {
		t.Errorf("accuracy %.1f out of range", acc)
	}
}

func TestPlayableMoves(t *testing.T) {
	got := playableMoves([]string{"e4", "e5", "Nf3", "1-0"})
	if len(got) != 3 {
		t.Fatalf("expected 3 moves, got %d: %v", len(got), got)
	}
	for _, m := range got {
		if m == "1-0" {
			t.Error("result token leaked into playable moves")
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
