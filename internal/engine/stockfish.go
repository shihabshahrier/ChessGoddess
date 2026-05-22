// Package engine drives a UCI chess engine (Stockfish) for position evaluation.
package engine

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MateScore is the centipawn magnitude used to represent a forced mate.
const MateScore = 100000

// Line is one engine principal variation, scored from the side-to-move POV.
type Line struct {
	Rank    int      `json:"rank"`     // MultiPV rank, 1 = best
	Depth   int      `json:"depth"`    // search depth reached
	ScoreCP int      `json:"score_cp"` // centipawns (0 when Mate != 0)
	Mate    int      `json:"mate"`     // moves to mate; sign = side-to-move POV; 0 if none
	PV      []string `json:"pv"`       // principal variation, UCI long-algebraic
}

// Score collapses cp/mate into one comparable centipawn value (side-to-move POV).
// A faster mate ranks above a slower one; a faster loss ranks below a slower one.
func (l Line) Score() int {
	switch {
	case l.Mate > 0:
		return MateScore - l.Mate
	case l.Mate < 0:
		return -MateScore - l.Mate
	default:
		return l.ScoreCP
	}
}

// Evaluation is the engine's verdict on a position.
type Evaluation struct {
	FEN      string `json:"fen"`
	Depth    int    `json:"depth"`
	BestMove string `json:"best_move"` // UCI long-algebraic
	Lines    []Line `json:"lines"`     // sorted by rank; Lines[0] is best
}

// Engine is a single Stockfish process. Evaluate is safe for serial use;
// concurrent callers should go through a Pool.
type Engine struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	mu     sync.Mutex
}

func New(stockfishPath string) (*Engine, error) {
	cmd := exec.Command(stockfishPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start stockfish: %w", err)
	}

	return &Engine{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}, nil
}

// Initialize completes the UCI handshake and waits for the engine to be ready.
func (e *Engine) Initialize() error {
	if err := e.sendCommand("uci"); err != nil {
		return err
	}
	if err := e.waitFor("uciok"); err != nil {
		return err
	}
	if err := e.sendCommand("isready"); err != nil {
		return err
	}
	return e.waitFor("readyok")
}

// Evaluate searches a position to the given depth, returning up to multipv lines.
func (e *Engine) Evaluate(fen string, depth, multipv int) (*Evaluation, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if depth < 1 {
		depth = 12
	}
	if multipv < 1 {
		multipv = 1
	}

	if err := e.sendCommand(fmt.Sprintf("setoption name MultiPV value %d", multipv)); err != nil {
		return nil, err
	}
	if err := e.sendCommand("position fen " + fen); err != nil {
		return nil, err
	}
	if err := e.sendCommand(fmt.Sprintf("go depth %d", depth)); err != nil {
		return nil, err
	}

	lines := make(map[int]Line)
	var bestMove string
	done := make(chan error, 1)

	go func() {
		for {
			raw, err := e.stdout.ReadString('\n')
			if err != nil {
				done <- err
				return
			}
			text := strings.TrimSpace(raw)
			if strings.HasPrefix(text, "bestmove") {
				if f := strings.Fields(text); len(f) >= 2 {
					bestMove = f[1]
				}
				done <- nil
				return
			}
			if strings.HasPrefix(text, "info") && strings.Contains(text, " pv ") {
				if pl := parseInfoLine(text); pl != nil {
					lines[pl.Rank] = *pl
				}
			}
		}
	}()

	select {
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("engine read failed: %w", err)
		}
	case <-time.After(90 * time.Second):
		_ = e.sendCommand("stop")
		if err := <-done; err != nil {
			return nil, fmt.Errorf("engine evaluation timed out: %w", err)
		}
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("engine returned no lines for %q", fen)
	}

	eval := &Evaluation{FEN: fen, BestMove: bestMove}
	for _, l := range lines {
		eval.Lines = append(eval.Lines, l)
	}
	sort.Slice(eval.Lines, func(i, j int) bool {
		return eval.Lines[i].Rank < eval.Lines[j].Rank
	})
	eval.Depth = eval.Lines[0].Depth
	return eval, nil
}

func (e *Engine) Close() error {
	_ = e.sendCommand("quit")
	return e.cmd.Wait()
}

func (e *Engine) sendCommand(cmd string) error {
	_, err := e.stdin.Write([]byte(cmd + "\n"))
	return err
}

func (e *Engine) waitFor(token string) error {
	for {
		text, err := e.stdout.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.TrimSpace(text) == token {
			return nil
		}
	}
}

// parseInfoLine extracts one PV line from a UCI "info ... pv ..." line.
// Returns nil for info lines that carry no principal variation.
func parseInfoLine(text string) *Line {
	fields := strings.Fields(text)
	line := &Line{Rank: 1}

	for i := 0; i < len(fields); i++ {
		switch fields[i] {
		case "depth":
			if i+1 < len(fields) {
				line.Depth, _ = strconv.Atoi(fields[i+1])
			}
		case "multipv":
			if i+1 < len(fields) {
				line.Rank, _ = strconv.Atoi(fields[i+1])
			}
		case "score":
			if i+2 < len(fields) {
				switch fields[i+1] {
				case "cp":
					line.ScoreCP, _ = strconv.Atoi(fields[i+2])
				case "mate":
					line.Mate, _ = strconv.Atoi(fields[i+2])
				}
			}
		case "pv":
			line.PV = append([]string(nil), fields[i+1:]...)
			if len(line.PV) == 0 {
				return nil
			}
			return line
		}
	}
	return nil
}
