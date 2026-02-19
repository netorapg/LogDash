package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/netorapg/LogDash/internal/core/parser"
	"github.com/netorapg/LogDash/internal/service"
)

// Model representa o estado da aplicação TUI
type Model struct {
	// Configuração
	rootPath string
	opts     service.AnalyzeOptions

	// Estado
	loading bool
	result  *service.AnalysisResult
	err     error

	// UI State
	activeTab int // 0: Overview, 1: Top Messages, 2: Timeline, 3: Anomalies
	width     int
	height    int

	// Mensagens de status
	statusMsg string
}

// Tabs disponíveis
const (
	tabOverview = iota
	tabTopMessages
	tabTimeline
	tabAnomalies
)

var tabNames = []string{"Overview", "Top Messages", "Timeline", "Anomalies"}

// NewModel cria novo modelo TUI
func NewModel(rootPath string, opts service.AnalyzeOptions) Model {
	if rootPath == "" {
		rootPath = "."
	}

	return Model{
		rootPath:  rootPath,
		opts:      opts,
		loading:   true,
		activeTab: tabOverview,
		width:     80,
		height:    24,
	}
}

// Init inicializa o modelo (comando inicial)
func (m Model) Init() tea.Cmd {
	return analyzeCmd(m.rootPath, m.opts)
}

// Update processa mensagens e atualiza o modelo
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "tab":
			// Próxima tab
			m.activeTab = (m.activeTab + 1) % len(tabNames)
			return m, nil

		case "shift+tab":
			// Tab anterior
			m.activeTab = (m.activeTab - 1 + len(tabNames)) % len(tabNames)
			return m, nil

		case "r":
			// Refresh - reanalizar
			m.loading = true
			m.statusMsg = "Refreshing..."
			return m, analyzeCmd(m.rootPath, m.opts)
		}

	case analysisResultMsg:
		m.loading = false
		m.result = msg.result
		m.err = msg.err
		if m.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", m.err)
		} else {
			m.statusMsg = "Analysis complete"
		}
		return m, nil
	}

	return m, nil
}

// View renderiza a UI
func (m Model) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	if m.result == nil {
		return "No data available. Press 'r' to refresh or 'q' to quit.\n"
	}

	// Header
	header := m.renderHeader()

	// Tabs
	tabs := m.renderTabs()

	// Content baseado na tab ativa
	var content string
	switch m.activeTab {
	case tabOverview:
		content = m.renderOverview()
	case tabTopMessages:
		content = m.renderTopMessages()
	case tabTimeline:
		content = m.renderTimeline()
	case tabAnomalies:
		content = m.renderAnomalies()
	}

	// Footer
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		footer,
	)
}

// Estilos
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

	tabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444"))

	activeTabStyle = lipgloss.NewStyle().
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B"))

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))
)

func (m Model) renderHeader() string {
	title := titleStyle.Render("LogDash TUI")
	path := dimStyle.Render(fmt.Sprintf("Path: %s", m.rootPath))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		path,
		"",
	)
}

func (m Model) renderTabs() string {
	tabs := []string{}

	for i, name := range tabNames {
		style := tabStyle
		if i == m.activeTab {
			style = activeTabStyle
		}
		tabs = append(tabs, style.Render(name))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n\n"
}

func (m Model) renderOverview() string {
	r := m.result

	var b strings.Builder

	// Summary
	b.WriteString(successStyle.Render("Summary") + "\n")
	b.WriteString(fmt.Sprintf("  %s\n\n", r.Summary))

	// Files
	b.WriteString(infoStyle.Render("Files") + "\n")
	b.WriteString(fmt.Sprintf("  Discovered: %d\n", r.FilesDiscovered))
	b.WriteString(fmt.Sprintf("  Processed:  %d\n", r.FilesProcessed))
	b.WriteString(fmt.Sprintf("  Entries:    %d\n\n", r.EntriesParsed))

	// By Level
	b.WriteString(infoStyle.Render("By Level") + "\n")

	if count := r.ByLevel[parser.LevelFatal]; count > 0 {
		b.WriteString(fmt.Sprintf("  💀 Fatal:   %d\n", count))
	}
	if count := r.ByLevel[parser.LevelError]; count > 0 {
		b.WriteString(errorStyle.Render(fmt.Sprintf("  ❌ Error:   %d\n", count)))
	}
	if count := r.ByLevel[parser.LevelWarning]; count > 0 {
		b.WriteString(warningStyle.Render(fmt.Sprintf("  ⚠️  Warning: %d\n", count)))
	}
	if count := r.ByLevel[parser.LevelInfo]; count > 0 {
		b.WriteString(fmt.Sprintf("  ℹ️  Info:    %d\n", count))
	}
	if count := r.ByLevel[parser.LevelDebug]; count > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  🐛 Debug:   %d\n", count)))
	}
	if count := r.ByLevel[parser.LevelTrace]; count > 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  🔍 Trace:   %d\n", count)))
	}

	return b.String()
}

func (m Model) renderTopMessages() string {
	r := m.result

	if len(r.TopMessages) == 0 {
		return dimStyle.Render("No messages found")
	}

	var b strings.Builder
	b.WriteString(infoStyle.Render("Top Messages") + "\n\n")

	for i, msg := range r.TopMessages {
		if i >= 15 { // Limitar a 15 para caber na tela
			break
		}

		icon := getLevelIcon(msg.Level)
		style := getLevelStyle(msg.Level)

		b.WriteString(fmt.Sprintf("%d. %s ", i+1, icon))
		b.WriteString(style.Render(fmt.Sprintf("[%s]", msg.Level)))
		b.WriteString(fmt.Sprintf(" %s ", truncate(msg.Message, 50)))
		b.WriteString(dimStyle.Render(fmt.Sprintf("(x%d)", msg.Count)))
		b.WriteString("\n")

		if !msg.FirstSeen.IsZero() {
			b.WriteString(dimStyle.Render(fmt.Sprintf("   First: %s, Last: %s\n",
				msg.FirstSeen.Format("15:04:05"),
				msg.LastSeen.Format("15:04:05"))))
		}
	}

	return b.String()
}

func (m Model) renderTimeline() string {
	r := m.result

	if len(r.Timeline) == 0 {
		return dimStyle.Render("No timeline data available")
	}

	var b strings.Builder
	b.WriteString(infoStyle.Render("Timeline") + "\n\n")

	// Encontrar max para escalar
	maxCount := 0
	for _, point := range r.Timeline {
		if point.Count > maxCount {
			maxCount = point.Count
		}
	}

	if maxCount == 0 {
		maxCount = 1
	}

	// Renderizar timeline
	for i, point := range r.Timeline {
		if i >= 20 { // Limitar a 20 pontos
			break
		}

		// Timestamp
		timestamp := point.Timestamp.Format("15:04")
		b.WriteString(dimStyle.Render(timestamp + " "))

		// Bar (escalar para máximo de 40 caracteres)
		barWidth := (point.Count * 40) / maxCount
		if barWidth == 0 && point.Count > 0 {
			barWidth = 1
		}

		bar := strings.Repeat("█", barWidth)
		b.WriteString(successStyle.Render(bar))

		// Count
		b.WriteString(fmt.Sprintf(" %d", point.Count))

		// Errors se houver
		if errors := point.ByLevel[parser.LevelError]; errors > 0 {
			b.WriteString(errorStyle.Render(fmt.Sprintf(" (%d errors)", errors)))
		}

		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderAnomalies() string {
	r := m.result

	if len(r.Anomalies) == 0 {
		return successStyle.Render("✓ No anomalies detected")
	}

	var b strings.Builder
	b.WriteString(warningStyle.Render(fmt.Sprintf("⚠️  %d Anomalies Detected", len(r.Anomalies))) + "\n\n")

	for i, anomaly := range r.Anomalies {
		if i >= 10 {
			break
		}

		// Severity icon
		icon := "ℹ️"
		style := infoStyle
		if anomaly.Severity >= 0.8 {
			icon = "🚨"
			style = errorStyle
		} else if anomaly.Severity >= 0.5 {
			icon = "⚠️"
			style = warningStyle
		}

		b.WriteString(fmt.Sprintf("%s ", icon))
		b.WriteString(style.Render(fmt.Sprintf("[%s]", anomaly.Type)))
		b.WriteString(fmt.Sprintf(" %s\n", anomaly.Description))
		b.WriteString(dimStyle.Render(fmt.Sprintf("   Severity: %.2f", anomaly.Severity)))
		if !anomaly.Timestamp.IsZero() {
			b.WriteString(dimStyle.Render(fmt.Sprintf(" | Time: %s", anomaly.Timestamp.Format("15:04:05"))))
		}
		b.WriteString("\n\n")
	}

	return b.String()
}

func (m Model) renderFooter() string {
	helpItems := []string{
		dimStyle.Render("tab/shift+tab: navigate"),
		dimStyle.Render("r: refresh"),
		dimStyle.Render("q: quit"),
	}

	help := strings.Join(helpItems, " • ")
	status := dimStyle.Render(fmt.Sprintf("Status: %s", m.statusMsg))

	return "\n" + lipgloss.JoinVertical(
		lipgloss.Left,
		strings.Repeat("─", m.width),
		help,
		status,
	)
}

func (m Model) renderLoading() string {
	return fmt.Sprintf("\n\n  🔍 Analyzing logs in %s...\n\n  Please wait...\n\n", m.rootPath)
}

func (m Model) renderError() string {
	return errorStyle.Render(fmt.Sprintf("\n\n  ❌ Error: %v\n\n  Press 'r' to retry or 'q' to quit.\n\n", m.err))
}

// Helper functions

func getLevelIcon(level parser.LogLevel) string {
	switch level {
	case parser.LevelFatal:
		return "💀"
	case parser.LevelError:
		return "❌"
	case parser.LevelWarning:
		return "⚠️"
	case parser.LevelInfo:
		return "ℹ️"
	case parser.LevelDebug:
		return "🐛"
	case parser.LevelTrace:
		return "🔍"
	default:
		return "  "
	}
}

func getLevelStyle(level parser.LogLevel) lipgloss.Style {
	switch level {
	case parser.LevelFatal, parser.LevelError:
		return errorStyle
	case parser.LevelWarning:
		return warningStyle
	case parser.LevelInfo:
		return infoStyle
	default:
		return dimStyle
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Commands

type analysisResultMsg struct {
	result *service.AnalysisResult
	err    error
}

func analyzeCmd(rootPath string, opts service.AnalyzeOptions) tea.Cmd {
	return func() tea.Msg {
		svc := service.NewLogDashService()
		opts.RootPath = rootPath

		result, err := svc.AnalyzeLogs(opts)

		return analysisResultMsg{
			result: result,
			err:    err,
		}
	}
}
