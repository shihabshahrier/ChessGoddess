package engine

import "testing"

func TestParseInfoLine_Centipawn(t *testing.T) {
	l := parseInfoLine("info depth 18 seldepth 24 multipv 1 score cp 55 nodes 100 pv e2e4 e7e5 g1f3")
	if l == nil {
		t.Fatal("expected a line, got nil")
	}
	if l.Depth != 18 || l.Rank != 1 || l.ScoreCP != 55 || l.Mate != 0 {
		t.Errorf("unexpected line: %+v", l)
	}
	if len(l.PV) != 3 || l.PV[0] != "e2e4" {
		t.Errorf("unexpected pv: %v", l.PV)
	}
}

func TestParseInfoLine_Mate(t *testing.T) {
	l := parseInfoLine("info depth 20 multipv 2 score mate -3 pv h7h8q")
	if l == nil {
		t.Fatal("expected a line, got nil")
	}
	if l.Mate != -3 || l.Rank != 2 {
		t.Errorf("unexpected line: %+v", l)
	}
}

func TestParseInfoLine_NoPV(t *testing.T) {
	if l := parseInfoLine("info depth 1 score cp 0 nodes 20"); l != nil {
		t.Errorf("expected nil for info line without pv, got %+v", l)
	}
}

func TestLineScore(t *testing.T) {
	if got := (Line{ScoreCP: 120}).Score(); got != 120 {
		t.Errorf("cp score = %d, want 120", got)
	}
	if (Line{Mate: 1}).Score() <= (Line{Mate: 5}).Score() {
		t.Error("a faster mate should score higher")
	}
	if (Line{Mate: -1}).Score() >= (Line{Mate: -5}).Score() {
		t.Error("a faster loss should score lower")
	}
	if (Line{Mate: 1}).Score() <= (Line{ScoreCP: 9000}).Score() {
		t.Error("a mate should outrank any centipawn score")
	}
}
