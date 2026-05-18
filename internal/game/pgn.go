package game

import (
	"regexp"
	"strings"
)

type PGN struct {
	Headers     map[string]string
	Moves       []string
	Result      string
	WhitePlayer string
	BlackPlayer string
	Event       string
	Date        string
	Opening     string
}

var headerRegex = regexp.MustCompile(`\[(\w+)\s+"(.+)"\]`)

func ParsePGN(pgn string) (*PGN, error) {
	pgn = strings.TrimSpace(pgn)
	
	headers := make(map[string]string)
	lines := strings.Split(pgn, "\n")
	
	var moveLines []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "" {
			continue
		}
		
		matches := headerRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			headers[matches[1]] = matches[2]
		} else {
			moveLines = append(moveLines, line)
		}
	}
	
	moves, result := parseMoves(strings.Join(moveLines, " "))
	
	pgnResult := headers["Result"]
	if pgnResult == "" && result != "" {
		pgnResult = result
	}
	
	return &PGN{
		Headers:     headers,
		Moves:       moves,
		Result:      pgnResult,
		WhitePlayer: headers["White"],
		BlackPlayer: headers["Black"],
		Event:       headers["Event"],
		Date:        headers["Date"],
		Opening:     headers["ECO"],
	}, nil
}

func parseMoves(moveText string) ([]string, string) {
	moveText = strings.TrimSpace(moveText)
	
	moveText = regexp.MustCompile(`\d+\.`).ReplaceAllString(moveText, "")
	
	moveText = regexp.MustCompile(`\{[^}]*\}`).ReplaceAllString(moveText, "")
	
	moveText = regexp.MustCompile(`;.*`).ReplaceAllString(moveText, "")
	
	resultRegex := regexp.MustCompile(`(1-0|0-1|1/2-1/2|\*)`)
	result := resultRegex.FindString(moveText)
	moveText = resultRegex.ReplaceAllString(moveText, "")
	
	moves := strings.Fields(moveText)
	
	var cleanMoves []string
	for _, move := range moves {
		move = strings.TrimSpace(move)
		if move != "" && move != "*" {
			cleanMoves = append(cleanMoves, move)
		}
	}
	
	return cleanMoves, result
}

func (p *PGN) GetHeader(key string) string {
	if val, ok := p.Headers[key]; ok {
		return val
	}
	return ""
}
