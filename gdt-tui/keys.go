package main

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

func handleKey(msg tea.KeyMsg, m *model) tea.Cmd {
	switch msg.String() {
	case "f10":
		m.activeTab = 0

	case "q", "Q":
		if m.state == stateIdle || m.state == stateDone || m.state == stateError {
			return tea.Quit
		}

	case "u", "U":
		if m.state == stateIdle {
			return cmdStartNgdt()
		}

	case "v", "V":
		if m.state == stateIdle {
			tunnelOnly = true
			return cmdStartNgdt()
		}

	case "r", "R":
		if m.state == stateDone {
			return cmdReboot()
		}
	}
	return nil
}

func cmdStartNgdt() tea.Cmd {
	return func() tea.Msg {
		go runNgdt(prog)
		return msgStateChange{state: stateConnecting}
	}
}

func cmdReboot() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("sudo", "reboot")
		_ = cmd.Start()
		return tea.Quit()
	}
}
