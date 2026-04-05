package main

import tea "github.com/charmbracelet/bubbletea"

type ngdtState int

const (
	stateIdle       ngdtState = iota
	stateConnecting           // session started
	stateTunnel               // tunnel active
	stateUpdating             // running steamos-update
	stateDone                 // update completed
	stateError                // error
)

func (s ngdtState) String() string {
	switch s {
	case stateIdle:
		return "idle"
	case stateConnecting:
		return "connecting"
	case stateTunnel:
		return "tunnel"
	case stateUpdating:
		return "updating"
	case stateDone:
		return "done"
	case stateError:
		return "error"
	}
	return "unknown"
}

type model struct {
	activeTab int
	state     ngdtState
	phase     string
	percent   int
	country   string
	sessionID string
	logs      []string
	notify    string
	width     int
	height    int
}

const maxLogs = 20

func initialModel() model {
	m := model{
		activeTab: 0,
		state:     stateIdle,
	}
	m.addLog(label("GDT · Гиккомовские инструменты для дека", "GDT · Geekcom tools for Deck"))
	return m
}

func (m *model) addLog(line string) {
	m.logs = append(m.logs, line)
	if len(m.logs) > maxLogs {
		m.logs = m.logs[len(m.logs)-maxLogs:]
	}
}

// --- tea.Msg types ---

type msgStateChange struct{ state ngdtState }
type msgPhase struct{ phase string }
type msgPercent struct{ percent int }
type msgCountry struct{ country string }
type msgSessionID struct{ id string }
type msgLog struct{ line string }
type msgNotify struct{ text string }
type msgNgdtDone struct{ err error }

// --- Update ---

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		cmd := handleKey(msg, &m)
		if cmd != nil {
			return m, cmd
		}

	case msgStateChange:
		m.state = msg.state

	case msgPhase:
		m.phase = msg.phase

	case msgPercent:
		m.percent = msg.percent

	case msgCountry:
		m.country = msg.country

	case msgSessionID:
		m.sessionID = msg.id

	case msgLog:
		m.addLog(msg.line)

	case msgNotify:
		m.notify = msg.text

	case msgNgdtDone:
		if msg.err != nil {
			m.state = stateError
			m.addLog("error: " + msg.err.Error())
		}
	}

	return m, nil
}

func (m model) View() string {
	return viewConsole(m)
}
