package game

import (
	"testing"
)

func TestParsePGN_BasicGame(t *testing.T) {
	pgn := `[Event "Test Game"]
[White "Player1"]
[Black "Player2"]
[Result "1-0"]

1. e4 e5 2. Nf3 Nc6 3. Bb5 a6 1-0`

	result, err := ParsePGN(pgn)
	if err != nil {
		t.Fatalf("failed to parse PGN: %v", err)
	}

	if result.WhitePlayer != "Player1" {
		t.Errorf("expected white player 'Player1', got '%s'", result.WhitePlayer)
	}

	if result.BlackPlayer != "Player2" {
		t.Errorf("expected black player 'Player2', got '%s'", result.BlackPlayer)
	}

	if result.Result != "1-0" {
		t.Errorf("expected result '1-0', got '%s'", result.Result)
	}

	if len(result.Moves) != 6 {
		t.Errorf("expected 6 moves, got %d", len(result.Moves))
	}
}

func TestParsePGN_NoHeaders(t *testing.T) {
	pgn := `1. e4 e5 2. Nf3 Nc6 1-0`

	result, err := ParsePGN(pgn)
	if err != nil {
		t.Fatalf("failed to parse PGN: %v", err)
	}

	if len(result.Moves) < 4 {
		t.Errorf("expected at least 4 moves, got %d", len(result.Moves))
	}
}

func TestParsePGN_DrawResult(t *testing.T) {
	pgn := `1. e4 e5 1/2-1/2`

	result, err := ParsePGN(pgn)
	if err != nil {
		t.Fatalf("failed to parse PGN: %v", err)
	}

	if result.Result != "1/2-1/2" {
		t.Errorf("expected result '1/2-1/2', got '%s'", result.Result)
	}
}

func TestParsePGN_BlackWin(t *testing.T) {
	pgn := `1. e4 e5 0-1`

	result, err := ParsePGN(pgn)
	if err != nil {
		t.Fatalf("failed to parse PGN: %v", err)
	}

	if result.Result != "0-1" {
		t.Errorf("expected result '0-1', got '%s'", result.Result)
	}
}

func TestParsePGN_GetHeader(t *testing.T) {
	pgn := `[Event "Test Event"]
[Site "Test Site"]

1. e4 e5 *`

	result, err := ParsePGN(pgn)
	if err != nil {
		t.Fatalf("failed to parse PGN: %v", err)
	}

	if result.GetHeader("Event") != "Test Event" {
		t.Errorf("expected event 'Test Event', got '%s'", result.GetHeader("Event"))
	}

	if result.GetHeader("Missing") != "" {
		t.Errorf("expected empty string for missing header, got '%s'", result.GetHeader("Missing"))
	}
}

func TestParsePGN_Comments(t *testing.T) {
	pgn := `1. e4 {King's Pawn} e5 {Open Game} 1-0`

	result, err := ParsePGN(pgn)
	if err != nil {
		t.Fatalf("failed to parse PGN: %v", err)
	}

	for _, move := range result.Moves {
		if move == "{King's Pawn}" || move == "{Open Game}" {
			t.Errorf("comment should be removed from moves, got '%s'", move)
		}
	}
}
