package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// runNgdt launches ./ngdt update, streams stdout/stderr line-by-line,
// parses known patterns and sends bubbletea messages via p.Send.
func runNgdt(p *tea.Program) {
	args := []string{}
	if tunnelOnly {
		args = append(args, "-n")
	} else {
		args = append(args, "update")
	}
	if serverFlag != "" {
		args = append(args, "-server", serverFlag)
	}
	cmd := exec.Command("./ngdt", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		p.Send(msgLog{line: fmt.Sprintf("stdout pipe: %v", err)})
		p.Send(msgNgdtDone{err: err})
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		p.Send(msgLog{line: fmt.Sprintf("stderr pipe: %v", err)})
		p.Send(msgNgdtDone{err: err})
		return
	}

	if err := cmd.Start(); err != nil {
		p.Send(msgLog{line: fmt.Sprintf("start: %v", err)})
		p.Send(msgNgdtDone{err: err})
		return
	}

	done := make(chan struct{}, 2)
	scan := func(s *bufio.Scanner) {
		for s.Scan() {
			line := s.Text()
			p.Send(msgLog{line: line})
			parseLine(p, line)
		}
		done <- struct{}{}
	}

	go scan(bufio.NewScanner(stdout))
	go scan(bufio.NewScanner(stderr))

	<-done
	<-done

	runErr := cmd.Wait()
	if runErr != nil {
		p.Send(msgLog{line: fmt.Sprintf("exit: %v", runErr)})
	}
	p.Send(msgNgdtDone{err: runErr})
}

// parseLine maps known output patterns to state/progress messages.
func parseLine(p *tea.Program, line string) {
	lower := strings.ToLower(line)

	switch {
	case strings.Contains(lower, "session started"):
		p.Send(msgStateChange{state: stateConnecting})

	case strings.Contains(lower, "tunnel active"):
		p.Send(msgStateChange{state: stateTunnel})
		// try to parse country: "tunnel active via DE"
		if idx := strings.Index(lower, "via "); idx != -1 {
			country := strings.TrimSpace(line[idx+4:])
			if sp := strings.IndexAny(country, " \t\r\n"); sp != -1 {
				country = country[:sp]
			}
			if country != "" {
				p.Send(msgCountry{country: country})
			}
		}

	case strings.Contains(lower, "running steamos-update"),
		strings.Contains(lower, "update available"):
		p.Send(msgStateChange{state: stateUpdating})
		p.Send(msgPhase{phase: "downloading update"})

	case strings.Contains(lower, "update completed successfully"),
		strings.Contains(lower, "update completed"):
		p.Send(msgStateChange{state: stateDone})
		p.Send(msgPhase{phase: "done"})
		p.Send(msgPercent{percent: 100})

	default:
		// Parse "42%" anywhere in the line
		for _, field := range strings.Fields(line) {
			trimmed := strings.TrimRight(field, "%")
			if trimmed == field {
				continue // no trailing %
			}
			if n, err := strconv.Atoi(trimmed); err == nil && n >= 0 && n <= 100 {
				p.Send(msgPercent{percent: n})
			}
		}
	}
}
