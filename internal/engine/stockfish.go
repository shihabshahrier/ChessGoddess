package engine

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MoveEvaluation struct {
	Move     string  `json:"move"`
	SAN      string  `json:"san"`
	FEN      string  `json:"fen"`
	Eval     float64 `json:"eval"`
	BestMove string  `json:"best_move"`
	Depth    int     `json:"depth"`
}

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

func (e *Engine) Initialize() error {
	if err := e.sendCommand("uci"); err != nil {
		return err
	}
	
	if err := e.waitForReady(); err != nil {
		return err
	}
	
	return e.sendCommand("isready")
}

func (e *Engine) SetPosition(fen string) error {
	return e.sendCommand(fmt.Sprintf("position fen %s", fen))
}

func (e *Engine) Analyze(depth int) ([]MoveEvaluation, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if err := e.sendCommand(fmt.Sprintf("go depth %d", depth)); err != nil {
		return nil, err
	}
	
	var evaluations []MoveEvaluation
	bestMove := ""
	var eval float64
	
	for {
		line, err := e.stdout.ReadString('\n')
		if err != nil {
			break
		}
		
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "bestmove") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				bestMove = parts[1]
			}
			break
		}
		
		if strings.HasPrefix(line, "info") && strings.Contains(line, "score") {
			eval = parseEvaluation(line)
			depth := parseDepth(line)
			pv := parsePV(line)
			
			if pv != "" {
				evaluations = append(evaluations, MoveEvaluation{
					BestMove: pv,
					Eval:     eval,
					Depth:    depth,
				})
			}
		}
	}
	
	if len(evaluations) > 0 {
		evaluations[len(evaluations)-1].BestMove = bestMove
	}
	
	return evaluations, nil
}

func (e *Engine) AnalyzeMove(fen, move string, depth int) (*MoveEvaluation, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if err := e.SetPosition(fen); err != nil {
		return nil, err
	}
	
	if err := e.sendCommand(fmt.Sprintf("go movetime %d", 1000)); err != nil {
		return nil, err
	}
	
	var bestMove string
	var eval float64
	var bestDepth int
	
	timeout := time.After(5 * time.Second)
	done := make(chan bool)
	
	go func() {
		for {
			line, err := e.stdout.ReadString('\n')
			if err != nil {
				done <- false
				return
			}
			
			line = strings.TrimSpace(line)
			
			if strings.HasPrefix(line, "bestmove") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					bestMove = parts[1]
				}
				done <- true
				return
			}
			
			if strings.HasPrefix(line, "info") && strings.Contains(line, "score") {
				eval = parseEvaluation(line)
				bestDepth = parseDepth(line)
			}
		}
	}()
	
	select {
	case <-timeout:
		return nil, fmt.Errorf("analysis timeout")
	case success := <-done:
		if !success {
			return nil, fmt.Errorf("analysis failed")
		}
	}
	
	return &MoveEvaluation{
		Move:     move,
		FEN:      fen,
		Eval:     eval,
		BestMove: bestMove,
		Depth:    bestDepth,
	}, nil
}

func (e *Engine) sendCommand(cmd string) error {
	_, err := e.stdin.Write([]byte(cmd + "\n"))
	return err
}

func (e *Engine) waitForReady() error {
	for {
		line, err := e.stdout.ReadString('\n')
		if err != nil {
			return err
		}
		
		if strings.TrimSpace(line) == "uciok" {
			return nil
		}
	}
}

func (e *Engine) Close() error {
	e.sendCommand("quit")
	return e.cmd.Wait()
}

func parseEvaluation(line string) float64 {
	parts := strings.Split(line, "score ")
	if len(parts) < 2 {
		return 0
	}
	
	scorePart := parts[1]
	
	if strings.HasPrefix(scorePart, "cp ") {
		cpStr := strings.Fields(scorePart)[1]
		cp, err := strconv.Atoi(cpStr)
		if err != nil {
			return 0
		}
		return float64(cp) / 100.0
	}
	
	if strings.HasPrefix(scorePart, "mate ") {
		mateStr := strings.Fields(scorePart)[1]
		mate, err := strconv.Atoi(mateStr)
		if err != nil {
			return 0
		}
		if mate > 0 {
			return 100.0
		}
		return -100.0
	}
	
	return 0
}

func parseDepth(line string) int {
	parts := strings.Split(line, "depth ")
	if len(parts) < 2 {
		return 0
	}
	
	depthStr := strings.Fields(parts[1])[0]
	depth, err := strconv.Atoi(depthStr)
	if err != nil {
		return 0
	}
	
	return depth
}

func parsePV(line string) string {
	parts := strings.Split(line, " pv ")
	if len(parts) < 2 {
		return ""
	}
	
	pv := strings.Fields(parts[1])
	if len(pv) > 0 {
		return pv[0]
	}
	
	return ""
}
