package engine

import (
	"testing"
)

func TestParseEvaluation_Centipawn(t *testing.T) {
	line := "info depth 10 seldepth 15 score cp 50 pv e4 e5"
	
	eval := parseEvaluation(line)
	
	if eval != 0.50 {
		t.Errorf("expected 0.50, got %f", eval)
	}
}

func TestParseEvaluation_NegativeCentipawn(t *testing.T) {
	line := "info depth 10 seldepth 15 score cp -30 pv e4 e5"
	
	eval := parseEvaluation(line)
	
	if eval != -0.30 {
		t.Errorf("expected -0.30, got %f", eval)
	}
}

func TestParseEvaluation_MatePositive(t *testing.T) {
	line := "info depth 20 seldepth 25 score mate 3 pv Qh5"
	
	eval := parseEvaluation(line)
	
	if eval != 100.0 {
		t.Errorf("expected 100.0, got %f", eval)
	}
}

func TestParseEvaluation_MateNegative(t *testing.T) {
	line := "info depth 20 seldepth 25 score mate -2 pv Qh5"
	
	eval := parseEvaluation(line)
	
	if eval != -100.0 {
		t.Errorf("expected -100.0, got %f", eval)
	}
}

func TestParseEvaluation_Zero(t *testing.T) {
	line := "info depth 10 seldepth 15 score cp 0 pv e4"
	
	eval := parseEvaluation(line)
	
	if eval != 0.0 {
		t.Errorf("expected 0.0, got %f", eval)
	}
}

func TestParseDepth(t *testing.T) {
	line := "info depth 15 seldepth 20 score cp 30 pv e4"
	
	depth := parseDepth(line)
	
	if depth != 15 {
		t.Errorf("expected depth 15, got %d", depth)
	}
}

func TestParsePV(t *testing.T) {
	line := "info depth 10 score cp 50 pv e4 e5 Nf3 Nc6"
	
	pv := parsePV(line)
	
	if pv != "e4" {
		t.Errorf("expected 'e4', got '%s'", pv)
	}
}

func TestParsePV_Empty(t *testing.T) {
	line := "info depth 10 score cp 50"
	
	pv := parsePV(line)
	
	if pv != "" {
		t.Errorf("expected empty string, got '%s'", pv)
	}
}
