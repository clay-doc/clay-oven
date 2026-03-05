package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Messages sent from the pipeline goroutine to the Bubble Tea runtime.
// ---------------------------------------------------------------------------

type logMsg struct {
	kind string // "success", "error", "warn", "info", "header", "keyval", "struct", "verbose"
	text string
	key  string // only for keyval
}

type confirmMsg struct {
	prompt string
}

type downloadProgressMsg struct {
	received int64
	total    int64
}

type pipelineDoneMsg struct{}

// ---------------------------------------------------------------------------
// TUISink – sends messages to the Bubble Tea program.
// ---------------------------------------------------------------------------

// TUISink implements OutputSink by sending tea messages to the TUI.
type TUISink struct {
	program   *tea.Program
	confirmMu sync.Mutex
	confirmCh chan bool
}

func NewTUISink() *TUISink {
	return &TUISink{
		confirmCh: make(chan bool, 1),
	}
}

func (t *TUISink) SetProgram(p *tea.Program) {
	t.program = p
}

func (t *TUISink) Banner() {
	// Banner is rendered by the TUI View(), nothing to send.
}

func (t *TUISink) Header(msg string) {
	t.program.Send(logMsg{kind: "header", text: msg})
}

func (t *TUISink) Success(msg string) {
	t.program.Send(logMsg{kind: "success", text: msg})
}

func (t *TUISink) Error(msg string) {
	t.program.Send(logMsg{kind: "error", text: msg})
}

func (t *TUISink) Warn(msg string) {
	t.program.Send(logMsg{kind: "warn", text: msg})
}

func (t *TUISink) Info(msg string) {
	t.program.Send(logMsg{kind: "info", text: msg})
}

func (t *TUISink) KeyVal(key, val string) {
	t.program.Send(logMsg{kind: "keyval", text: val, key: key})
}

func (t *TUISink) StructLine(line string) {
	t.program.Send(logMsg{kind: "struct", text: line})
}

func (t *TUISink) Verbose(msg string) {
	if Verbose {
		t.program.Send(logMsg{kind: "verbose", text: msg})
	}
}

// Confirm blocks the pipeline goroutine until the user presses y/n in the TUI.
func (t *TUISink) Confirm(prompt string) bool {
	t.confirmMu.Lock()
	defer t.confirmMu.Unlock()

	t.program.Send(confirmMsg{prompt: prompt})
	return <-t.confirmCh
}

func (t *TUISink) DownloadProgress(received, total int64) {
	t.program.Send(downloadProgressMsg{received: received, total: total})
}

func (t *TUISink) Done() {
	t.program.Send(pipelineDoneMsg{})
}

// ---------------------------------------------------------------------------
// Bubble Tea Model
// ---------------------------------------------------------------------------

type tuiModel struct {
	logs     []logMsg
	spinner  spinner.Model
	progress progress.Model
	sink     *TUISink

	// Confirm state
	confirmPrompt  string
	confirmPending bool

	// Download state
	dlReceived int64
	dlTotal    int64
	dlActive   bool

	// Pipeline state
	done   bool
	width  int
	height int
}

func newTUIModel(sink *TUISink) tuiModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(colorCyan)

	prog := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return tuiModel{
		spinner:  sp,
		progress: prog,
		sink:     sink,
	}
}

func (m tuiModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.confirmPending {
				m.sink.confirmCh <- false
				m.confirmPending = false
				return m, tea.Quit
			}
			return m, tea.Quit

		case "y", "Y":
			if m.confirmPending {
				m.confirmPending = false
				m.logs = append(m.logs, logMsg{kind: "success", text: fmt.Sprintf("%s → yes", m.confirmPrompt)})
				m.sink.confirmCh <- true
				return m, nil
			}

		case "n", "N":
			if m.confirmPending {
				m.confirmPending = false
				m.logs = append(m.logs, logMsg{kind: "warn", text: fmt.Sprintf("%s → no", m.confirmPrompt)})
				m.sink.confirmCh <- false
				return m, nil
			}
		}

	case logMsg:
		m.logs = append(m.logs, msg)
		return m, nil

	case confirmMsg:
		m.confirmPending = true
		m.confirmPrompt = msg.prompt
		return m, nil

	case downloadProgressMsg:
		m.dlActive = true
		m.dlReceived = msg.received
		m.dlTotal = msg.total
		if msg.total > 0 && msg.received >= msg.total {
			m.dlActive = false
		}
		return m, nil

	case pipelineDoneMsg:
		m.done = true
		m.dlActive = false
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m tuiModel) View() string {
	var b strings.Builder

	// Banner
	banner := `
   ╔═══════════════════════════════════════╗
   ║           C L A Y   O V E N           ║
   ╚═══════════════════════════════════════╝`
	b.WriteString(styleBanner.Render(banner))
	b.WriteString("\n\n")

	// Log lines
	for _, entry := range m.logs {
		switch entry.kind {
		case "header":
			bar := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
			b.WriteString("\n")
			b.WriteString(styleDim.Render(bar))
			b.WriteString("\n")
			b.WriteString(styleHeader.Render("▸ " + entry.text))
			b.WriteString("\n")
			b.WriteString(styleDim.Render(bar))
			b.WriteString("\n")
		case "success":
			b.WriteString(styleSuccess.Render("  " + iconSuccess + " " + entry.text))
			b.WriteString("\n")
		case "error":
			b.WriteString(styleError.Render("  " + iconError + " " + entry.text))
			b.WriteString("\n")
		case "warn":
			b.WriteString(styleWarn.Render("  " + iconWarn + " " + entry.text))
			b.WriteString("\n")
		case "info":
			b.WriteString(styleInfo.Render("  " + iconInfo + " " + entry.text))
			b.WriteString("\n")
		case "keyval":
			formatted := fmt.Sprintf("  %-20s %s %s",
				styleKey.Render(entry.key),
				styleDim.Render(iconArrow),
				styleVal.Render(entry.text))
			b.WriteString(formatted)
			b.WriteString("\n")
		case "struct":
			line := strings.TrimRight(entry.text, "\n")
			b.WriteString(styleStructLine.Render(line))
			b.WriteString("\n")
		case "verbose":
			b.WriteString(styleDim.Render("  " + iconDot + " " + entry.text))
			b.WriteString("\n")
		}
	}

	// Download progress bar
	if m.dlActive && m.dlTotal > 0 {
		pct := float64(m.dlReceived) / float64(m.dlTotal)
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("  %s Downloading... %s",
			m.spinner.View(),
			m.progress.ViewAs(pct)))
		b.WriteString("\n")
	} else if m.dlActive {
		b.WriteString(fmt.Sprintf("\n  %s Downloading... %d bytes",
			m.spinner.View(), m.dlReceived))
		b.WriteString("\n")
	}

	// Confirm prompt
	if m.confirmPending {
		b.WriteString("\n")
		prompt := fmt.Sprintf("  %s %s ",
			styleWarn.Render("?"),
			styleInfo.Render(m.confirmPrompt))
		hint := styleDim.Render("(y/n)")
		b.WriteString(prompt + hint)
		b.WriteString("\n")
	}

	// Spinner when not done and not confirming
	if !m.done && !m.confirmPending && !m.dlActive {
		b.WriteString(fmt.Sprintf("\n  %s Working...\n", m.spinner.View()))
	}

	// Done message
	if m.done {
		b.WriteString("\n")
		b.WriteString(styleSuccess.Render("  " + iconSuccess + " All done!"))
		b.WriteString("\n")
	}

	return b.String()
}
