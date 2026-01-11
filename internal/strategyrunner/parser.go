package strategyrunner

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

// Parser parses .bat strategy files into internal representation.
type Parser struct {
	variables      map[string]string
	gameFilter     bool
	gameFilterPorts string
	logger         *slog.Logger
}

// ParsedStrategy represents a parsed strategy with rules.
type ParsedStrategy struct {
	Rules []ParsedRule
}

// ParsedRule represents a single parsed rule.
type ParsedRule struct {
	// Protocol is "tcp" or "udp"
	Protocol string

	// Ports is a comma-separated list of ports or ranges
	Ports string

	// NFQWSArgs contains all arguments for nfqws
	NFQWSArgs string

	// QueueNum is the sequential queue number
	QueueNum int
}

// NewParser creates a new BAT file parser.
func NewParser(binPath, listsPath, gameFilterPorts string, gameFilterEnabled bool, logger *slog.Logger) *Parser {
	return &Parser{
		variables: map[string]string{
			"BIN":        binPath,
			"LISTS":      listsPath,
			"GameFilter": gameFilterPorts,
		},
		gameFilter:      gameFilterEnabled,
		gameFilterPorts: gameFilterPorts,
		logger:          logger,
	}
}

// Parse parses a .bat strategy file.
func (p *Parser) Parse(filepath string) (*ParsedStrategy, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open strategy file: %w", err)
	}
	defer file.Close()

	var rules []ParsedRule
	queueNum := 0
	filterRegex := regexp.MustCompile(`--filter-(tcp|udp)=([0-9,-]+)\s+(.*?)(?:--new|$)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments and service lines
		if p.isSkipLine(line) {
			continue
		}

		// Apply variable substitution
		line = p.substituteVariables(line)

		// Find all filter rules in the line
		matches := filterRegex.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}

		for _, match := range matches {
			protocol := match[1]
			ports := match[2]
			nfqwsArgs := strings.TrimSpace(match[3])

			// Skip empty args
			if nfqwsArgs == "" {
				continue
			}

			// Clean up the args (remove quotes and leading dashes)
			nfqwsArgs = p.cleanArgs(nfqwsArgs)

			rule := ParsedRule{
				Protocol:  protocol,
				Ports:     ports,
				NFQWSArgs: nfqwsArgs,
				QueueNum:  queueNum,
			}

			p.logger.Debug("parsed rule",
				slog.String("protocol", protocol),
				slog.String("ports", ports),
				slog.Int("queue", queueNum),
			)

			rules = append(rules, rule)
			queueNum++
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading strategy file: %w", err)
	}

	if len(rules) == 0 {
		return nil, fmt.Errorf("no filter rules found in strategy file")
	}

	return &ParsedStrategy{Rules: rules}, nil
}

// isSkipLine checks if a line should be skipped.
func (p *Parser) isSkipLine(line string) bool {
	line = strings.TrimSpace(line)

	// Skip empty lines
	if line == "" {
		return true
	}

	// Skip comments
	if strings.HasPrefix(line, "::") || strings.HasPrefix(line, "@echo") || strings.HasPrefix(line, "rem ") {
		return true
	}

	// Skip service commands
	if strings.Contains(line, "chcp ") || strings.Contains(line, "cd /d ") ||
		strings.Contains(line, "call service.bat") || strings.Contains(line, "set \"BIN") ||
		strings.Contains(line, "set \"LISTS") {
		return true
	}

	// Skip lines without filter rules or useful content
	if !strings.Contains(line, "--filter-") && !strings.Contains(line, "--hostlist") &&
		!strings.Contains(line, "--ipset") && !strings.Contains(line, "--dpi-desync") {
		return true
	}

	return false
}

// substituteVariables replaces variables in a line.
func (p *Parser) substituteVariables(line string) string {
	// Replace %BIN%
	line = strings.ReplaceAll(line, "%BIN%", p.variables["BIN"])

	// Replace %LISTS%
	line = strings.ReplaceAll(line, "%LISTS%", p.variables["LISTS"])

	// Handle %GameFilter%
	if p.gameFilter {
		line = strings.ReplaceAll(line, "%GameFilter%", p.variables["GameFilter"])
	} else {
		// Remove GameFilter references when disabled
		// Remove ,%GameFilter% and %GameFilter%, and standalone %GameFilter%
		line = strings.ReplaceAll(line, ",%GameFilter%", "")
		line = strings.ReplaceAll(line, "%GameFilter%,", "")
		line = strings.ReplaceAll(line, "%GameFilter%", "")
		// Clean up double commas that might result
		for strings.Contains(line, ",,") {
			line = strings.ReplaceAll(line, ",,", ",")
		}
		// Clean up trailing/leading commas
		for strings.Contains(line, ",}") || strings.Contains(line, "{,") {
			line = strings.ReplaceAll(line, ",}", "}")
			line = strings.ReplaceAll(line, "{,", "{")
		}
	}

	// Handle line continuations (^ in batch files)
	line = strings.ReplaceAll(line, "^", "")

	return line
}

// cleanArgs cleans up nfqws arguments.
func (p *Parser) cleanArgs(args string) string {
	// Remove leading/trailing whitespace
	args = strings.TrimSpace(args)

	// Replace escaped quotes with regular quotes
	args = strings.ReplaceAll(args, `\"`, `"`)

	// Remove newlines and extra spaces
	args = strings.Join(strings.Fields(args), " ")

	return args
}
