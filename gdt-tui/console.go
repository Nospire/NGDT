package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ---- palette ---------------------------------------------------------------

const (
	clrBg        = lipgloss.Color("#0d0a07")
	clrAccent    = lipgloss.Color("#c47a3a")
	clrText      = lipgloss.Color("#c8b89a")
	clrDim       = lipgloss.Color("#5a4a38")
	clrDimDark   = lipgloss.Color("#4a3520") // hotkey descriptions
	clrGreen     = lipgloss.Color("#7ec47a")
	clrGreenDark = lipgloss.Color("#3d6b3d") // tunnel dot active
	clrOrange    = lipgloss.Color("#e09050")
	clrRed       = lipgloss.Color("#c45a3a")
	clrDotDim    = lipgloss.Color("#3d2e1a") // inactive dot
)

// ---- styles ----------------------------------------------------------------

var (
	styleBorder = lipgloss.NewStyle().
			Background(clrBg).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(clrAccent)

	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrAccent).
			Background(clrBg)

	styleTabActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrBg).
			Background(clrAccent).
			Padding(0, 1)

	styleTabInactive = lipgloss.NewStyle().
				Foreground(clrDim).
				Background(clrBg).
				Padding(0, 1)

	styleButton = lipgloss.NewStyle().
			Bold(true).
			Foreground(clrBg).
			Background(clrAccent).
			Padding(0, 1)

	styleButtonDisabled = lipgloss.NewStyle().
				Foreground(clrDim).
				Background(clrBg).
				Padding(0, 1)

	styleState = map[ngdtState]lipgloss.Style{
		stateIdle:       lipgloss.NewStyle().Foreground(clrDim).Background(clrBg),
		stateConnecting: lipgloss.NewStyle().Foreground(clrAccent).Background(clrBg),
		stateTunnel:     lipgloss.NewStyle().Foreground(clrAccent).Background(clrBg),
		stateUpdating:   lipgloss.NewStyle().Foreground(clrOrange).Background(clrBg),
		stateDone:       lipgloss.NewStyle().Foreground(clrGreen).Background(clrBg),
		stateError:      lipgloss.NewStyle().Foreground(clrRed).Background(clrBg),
	}

	styleLog  = lipgloss.NewStyle().Foreground(clrText).Background(clrBg)
	styleHint = lipgloss.NewStyle().Foreground(clrDim).Background(clrBg)
	stylePct  = lipgloss.NewStyle().Foreground(clrDim).Background(clrBg)
	styleFill = lipgloss.NewStyle().Background(clrBg)
)

// ---- helpers ---------------------------------------------------------------

// label returns the Russian string in normal terminals and the English
// fallback when running in a basic TTY (no Unicode support).
func label(ru, en string) string {
	if ttyMode {
		return en
	}
	return ru
}

func dot(color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(color).Background(clrBg).Render("●") + " "
}

func sectionLabel(s string) string {
	return lipgloss.NewStyle().Foreground(clrDim).Background(clrBg).Render(s)
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) > max {
		return string(r[:max]) + "…"
	}
	return s
}

// ---- tabs header -----------------------------------------------------------

func renderTabs(m model) string {
	tabs := styleTabActive.Render("Console")
	hint := styleHint.Render("F10  Q=quit")
	gap := m.width - lipgloss.Width(tabs) - lipgloss.Width(hint) - 2
	if gap < 1 {
		gap = 1
	}
	return styleFill.Width(m.width).Render(
		tabs + strings.Repeat(" ", gap) + hint,
	)
}

// ---- progress bar ----------------------------------------------------------

func renderProgressBar(pct int, width int) string {
	if width < 4 {
		width = 4
	}
	filled := int(float64(width) * float64(pct) / 100.0)
	if filled > width {
		filled = width
	}
	barColor := clrOrange
	if pct >= 100 {
		barColor = clrGreen
	}
	bar := lipgloss.NewStyle().Foreground(barColor).Background(clrBg).Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().Foreground(clrDim).Background(clrBg).Render(strings.Repeat("░", width-filled))
	return bar + stylePct.Render(fmt.Sprintf(" %3d%%", pct))
}

// ---- left (actions) column -------------------------------------------------

func renderActionsPane(m model, w, h int) string {
	keyStyle := lipgloss.NewStyle().Foreground(clrAccent).Background(clrBg).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(clrDimDark).Background(clrBg)

	// buttons
	updateBtn := styleButton.Render("[ U ] Update SteamOS")
	vpnBtn := styleButton.Render("[ V ] VPN only")
	if m.state != stateIdle {
		updateBtn = styleButtonDisabled.Render("[ U ] Update SteamOS")
		vpnBtn = ""
	}
	rebootBtn := ""
	if m.state == stateDone {
		rebootBtn = "\n" + styleButton.Render("[ R ] Reboot")
	}
	btns := updateBtn
	if vpnBtn != "" {
		btns += "\n" + vpnBtn
	}
	btns += rebootBtn

	actionsBlock := styleHeader.Render(label("ДЕЙСТВИЯ", "ACTIONS")) + "\n\n" + btns

	// hotkeys block
	type hk struct{ key, desc string }
	hotkeys := []hk{
		{"U", label("обновить", "update")},
		{"V", label("только VPN", "VPN only")},
		{"R", label("перезагрузка", "reboot")},
		{"Q", label("выйти", "quit")},
	}
	var hkLines []string
	for _, h := range hotkeys {
		hkLines = append(hkLines,
			keyStyle.Render(fmt.Sprintf("%-7s", h.key))+" "+descStyle.Render(h.desc),
		)
	}
	hotkeysBlock := styleHeader.Render(label("ХОТКЕИ", "HOTKEYS")) + "\n" + strings.Join(hkLines, "\n")

	// push hotkeys to bottom: fill lines between
	actH := strings.Count(actionsBlock, "\n") + 1
	hkH := strings.Count(hotkeysBlock, "\n") + 1
	spacerH := h - actH - hkH - 2
	if spacerH < 1 {
		spacerH = 1
	}
	spacer := strings.Repeat("\n", spacerH)

	return styleBorder.Width(w).Height(h).Render(
		actionsBlock + spacer + hotkeysBlock,
	)
}

// ---- right (status) column -------------------------------------------------

func renderStatusPane(m model, w, h int) string {
	st := m.state

	// --- ТУННЕЛЬ ---
	tunnelDot := dot(clrDotDim)
	tunnelVal := lipgloss.NewStyle().Foreground(clrDim).Background(clrBg).Render(label("не активен", "inactive"))
	if st >= stateTunnel {
		tunnelDot = dot(clrGreenDark)
		val := label("активен", "active")
		if m.country != "" {
			val += " · " + m.country
		}
		tunnelVal = lipgloss.NewStyle().Foreground(clrText).Background(clrBg).Render(val)
	}
	tunnelBlock := sectionLabel(label("ТУННЕЛЬ", "TUNNEL")) + "\n" +
		tunnelDot + tunnelVal

	// --- ОБНОВЛЕНИЕ ---
	updDot := dot(clrDotDim)
	updVal := lipgloss.NewStyle().Foreground(clrDim).Background(clrBg).Render(label("ожидание", "waiting"))
	if st == stateUpdating {
		updDot = dot(clrAccent)
		v := m.phase
		if m.percent > 0 {
			v += " · " + strconv.Itoa(m.percent) + "%"
		}
		updVal = lipgloss.NewStyle().Foreground(clrOrange).Background(clrBg).Render(v)
	} else if st == stateDone {
		updDot = dot(clrGreen)
		v := m.phase
		if m.percent > 0 {
			v += " · " + strconv.Itoa(m.percent) + "%"
		}
		updVal = lipgloss.NewStyle().Foreground(clrGreen).Background(clrBg).Render(v)
	}
	updateBlock := sectionLabel(label("ОБНОВЛЕНИЕ", "UPDATE")) + "\n" +
		updDot + updVal

	// --- СЕССИЯ ---
	sessID := m.sessionID
	if sessID == "" {
		sessID = "—"
	} else {
		sessID = truncate(sessID, 18)
	}
	sessIDLine := lipgloss.NewStyle().Foreground(clrText).Background(clrBg).Render(sessID)
	sessCountry := ""
	if m.country != "" {
		sessCountry = "\n" + lipgloss.NewStyle().Foreground(clrAccent).Background(clrBg).Render(m.country)
	}
	sessionBlock := sectionLabel(label("СЕССИЯ", "SESSION")) + "\n" + sessIDLine + sessCountry

	content := styleHeader.Render(label("СТАТУС", "STATUS")) + "\n\n" +
		tunnelBlock + "\n\n" +
		updateBlock + "\n\n" +
		sessionBlock

	return styleBorder.Width(w).Height(h).Render(content)
}

// ---- bottom bar ------------------------------------------------------------

func renderBottomBar(m model) string {
	if m.state == stateDone {
		return lipgloss.NewStyle().
			Width(m.width).
			Background(lipgloss.Color("#1a2e1a")).
			Foreground(lipgloss.Color("#4a9a4a")).
			Padding(0, 1).
			Render("● " + label("Обновление завершено · нажми R для перезагрузки", "Update done · press R to reboot"))
	}
	hints := label(
		"U обновить · V только VPN · R перезагрузка · Q выйти",
		"U update · V VPN only · R reboot · Q quit",
	)
	return lipgloss.NewStyle().
		Width(m.width).
		Background(lipgloss.Color("#1a1208")).
		Foreground(lipgloss.Color("#3d2e1a")).
		Padding(0, 1).
		Render(hints)
}

// ---- three-column console view ---------------------------------------------

func viewConsole(m model) string {
	if m.width == 0 {
		return "loading…"
	}

	actW := 24
	stW := 26
	logW := m.width - actW - stW - 6
	if logW < 20 {
		logW = 20
	}
	// tabs(1) + bottom bar(1) + borders(2) = 4 → leave 1 extra for bottom bar
	innerH := m.height - 5
	if innerH < 1 {
		innerH = 1
	}

	// ---- log + progress column ----
	logLines := make([]string, 0, len(m.logs))
	for _, l := range m.logs {
		runes := []rune(l)
		if len(runes) > logW-2 {
			l = string(runes[:logW-2])
		}
		logLines = append(logLines, styleLog.Render(l))
	}
	for len(logLines) < innerH-4 {
		logLines = append(logLines, "")
	}
	logsText := strings.Join(logLines, "\n")

	progressSection := ""
	if m.state == stateUpdating || m.state == stateDone {
		progressSection = "\n" + renderProgressBar(m.percent, logW-4)
		if m.phase != "" {
			progressSection += "\n" + styleHint.Render(m.phase)
		}
	}

	logPane := styleBorder.
		Width(logW).Height(innerH).
		Render(styleHeader.Render(label("ЛОГ", "LOG")) + "\n" + logsText + progressSection)

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		renderActionsPane(m, actW, innerH),
		logPane,
		renderStatusPane(m, stW, innerH),
	)
	return renderTabs(m) + "\n" + row + "\n" + renderBottomBar(m)
}
