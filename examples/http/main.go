package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/williamnoble/bubbletea-toast" // Use your actual import path
)

// Styles
var (
	appStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3c3836"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#83a598")).
			Bold(true).
			MarginBottom(1)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#83a598")).
			Padding(0, 1)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#282828")).
			Background(lipgloss.Color("#83a598")).
			Padding(0, 3).
			MarginTop(1)

	focusedButtonStyle = buttonStyle.
				Background(lipgloss.Color("#b8bb26"))
)

// Key mappings
type keyMap struct {
	Enter key.Binding
	Tab   key.Binding
	Ping  key.Binding
	Quit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Tab, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Enter, k.Tab},
		{k.Ping, k.Quit},
	}
}

var keys = keyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "submit"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	Ping: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "ping host"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// Model
type model struct {
	input        textinput.Model
	toastManager toast.Manager
	help         help.Model
	keys         keyMap
	focused      int
	width        int
	height       int
	loading      bool
}

func initialModel() model {
	input := textinput.New()
	input.Placeholder = "Enter hostname (e.g., google.com)"
	input.Focus()
	input.CharLimit = 156
	input.Width = 40

	return model{
		input:        input,
		toastManager: toast.NewManager().WithPosition(toast.BottomRight),
		help:         help.New(),
		keys:         keys,
		focused:      0,
		loading:      false,
	}
}

// Message types
type pingResultMsg struct {
	Host    string
	Success bool
	Message string
}

// Ping command
func pingCmd(host string) tea.Cmd {
	return func() tea.Msg {
		if host == "" {
			return pingResultMsg{
				Success: false,
				Message: "Please enter a hostname",
			}
		}

		// Clean the input a bit
		host = strings.TrimSpace(host)
		if !strings.Contains(host, ":") {
			host += ":80" // Default to port 80
		}

		start := time.Now()
		conn, err := net.DialTimeout("tcp", host, 3*time.Second)
		elapsed := time.Since(start)

		if err != nil {
			return pingResultMsg{
				Host:    host,
				Success: false,
				Message: fmt.Sprintf("Failed to connect: %v", err),
			}
		}
		defer conn.Close()

		return pingResultMsg{
			Host:    host,
			Success: true,
			Message: fmt.Sprintf("Connected in %s", elapsed),
		}
	}
}

// Init initializes the application
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Ignore key presses when loading
		if m.loading {
			break
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Ping) || key.Matches(msg, m.keys.Enter):
			m.loading = true
			cmds = append(cmds, pingCmd(m.input.Value()))
			cmds = append(cmds, m.toastManager.Push(
				"Pinging host...",
				toast.Info,
				4*time.Second,
			))

		case key.Matches(msg, m.keys.Tab):
			// Toggle focus between input and buttons
			m.focused = (m.focused + 1) % 2
			if m.focused == 0 {
				m.input.Focus()
			} else {
				m.input.Blur()
			}
		}

	case pingResultMsg:
		m.loading = false
		if msg.Success {
			cmds = append(cmds, m.toastManager.Push(
				fmt.Sprintf("✓ %s: %s", msg.Host, msg.Message),
				toast.Info,
				4*time.Second,
			))
		} else {
			cmds = append(cmds, m.toastManager.Push(
				fmt.Sprintf("✗ %s", msg.Message),
				toast.Error,
				4*time.Second,
			))
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.toastManager = m.toastManager.WithWidth(msg.Width / 3)
	}

	// update input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	// update toast manager
	m.toastManager, cmd = m.toastManager.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m model) View() string {
	if m.width == 0 {
		// Initial render before we know the terminal size
		return "Loading..."
	}

	// Title
	title := titleStyle.Render("Simple Ping Tool")

	// Input field
	inputField := inputStyle.Render(m.input.View())

	// Ping button
	var pingButton string
	if m.focused == 1 {
		pingButton = focusedButtonStyle.Render("[ Ping ]")
	} else {
		pingButton = buttonStyle.Render("[ Ping ]")
	}

	// Status text
	var status string
	if m.loading {
		status = "Connecting..."
	} else {
		status = "Ready"
	}

	// Main content area
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"Enter a hostname to ping:",
		inputField,
		pingButton,
		"",
		fmt.Sprintf("Status: %s", status),
	)

	// Help view
	helpView := m.help.View(m.keys)

	// Main layout
	mainView := lipgloss.JoinVertical(
		lipgloss.Left,
		appStyle.Render(content),
		"",
		helpView,
	)

	// Toast view
	toastsView := m.toastManager.View()
	if toastsView != "" {
		return lipgloss.JoinVertical(lipgloss.Left, mainView, toastsView)
	}

	return mainView
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
