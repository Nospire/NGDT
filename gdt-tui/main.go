package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// prog is set once in main so that ngdt.go can call prog.Send from goroutines.
var prog *tea.Program

// ttyMode is true when running in a basic Linux TTY (no Unicode/truecolor).
var ttyMode bool

var (
	localMode  bool   // -lo: no-op, kept for script compatibility
	serverFlag string // "NL", "PL", "BG", "LV"
	tunnelOnly bool
)

func isTTY() bool {
	term := os.Getenv("TERM")
	return term == "linux" || term == ""
}

func main() {
	// 0. Parse flags manually (avoid flag package conflicting with bubbletea).
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-lo":
			localMode = true // kept as no-op for compatibility
		case "-n":
			tunnelOnly = true
		case "-nl":
			serverFlag = "NL"
		case "-pl":
			serverFlag = "PL"
		case "-bg":
			serverFlag = "BG"
		case "-lv":
			serverFlag = "LV"
		}
	}

	// 1. Check that deck's password is set (passwd -S deck → second field must not be "L")
	out, err := exec.Command("passwd", "-S", "deck").Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: passwd -S deck: %v\n", err)
		os.Exit(1)
	}
	fields := strings.Fields(string(out))
	if len(fields) >= 2 && fields[1] == "L" {
		fmt.Fprintln(os.Stderr, "error: deck account has no password set (locked).")
		fmt.Fprintln(os.Stderr, "Set a password first:  passwd deck")
		os.Exit(1)
	}

	// 2. Prime sudo credentials — use GDT_SUDO_PASS from env if available.
	if pass := os.Getenv("GDT_SUDO_PASS"); pass != "" {
		sudoV := exec.Command("sudo", "-S", "-k", "-p", "", "-v")
		sudoV.Stdin = strings.NewReader(pass + "\n")
		if err := sudoV.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error: sudo -v failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		sudoV := exec.Command("sudo", "-v")
		sudoV.Stdin = os.Stdin
		sudoV.Stdout = os.Stdout
		sudoV.Stderr = os.Stderr
		if err := sudoV.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error: sudo -v failed: %v\n", err)
			os.Exit(1)
		}
	}

	// 3. Detect TTY mode before launching TUI.
	ttyMode = isTTY()

	// 4. Force truecolor so lipgloss hex colours work in all terminals.
	os.Setenv("COLORTERM", "truecolor")

	// 5. Launch the TUI.
	prog = tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := prog.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
