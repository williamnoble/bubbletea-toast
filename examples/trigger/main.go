package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	toast "github.com/williamnoble/bubbletea-toast"
	"os"
	"time"
)

var (
	mainContentStyle = lipgloss.NewStyle().
				Width(80).
				Height(10).
				Align(lipgloss.Center, lipgloss.Center).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#3c3836"))

	// custom Type
	PartyType toast.Type = 101

	partyStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			Width(30).
			Foreground(lipgloss.Color("#FFD700")).       // Gold text
			BorderForeground(lipgloss.Color("#FF1493")). // Deep pink border
			Italic(true)
)

type keyMap struct {
	Info       key.Binding
	Warning    key.Binding
	Error      key.Binding
	Quit       key.Binding
	DismissAll key.Binding

	// custom Type
	Party key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Info, k.Warning, k.Error, k.Quit, k.DismissAll, k.Party}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Info, k.Warning, k.Error, k.DismissAll, k.Party},
		{k.Quit},
	}
}

var keys = keyMap{
	Info: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "show info toast"),
	),
	Warning: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "show warning toast"),
	),
	Error: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "show error toast"),
	),
	Party: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "show party toast"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	DismissAll: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "dismiss all toasts"),
	),
}

type model struct {
	toastManager toast.Manager
	help         help.Model
	keys         keyMap
}

func initialModel() model {
	return model{
		toastManager: toast.NewManager().WithStyle(PartyType, "🎉", partyStyle),
		help:         help.New(),
		keys:         keys,
	}
}

func (m model) Init() tea.Cmd {
	return m.toastManager.Push(
		"Welcome to the toast demo!",
		toast.Info,
		1*time.Second)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var commands []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Info):
			commands = append(commands, m.toastManager.Push(
				"This is an info notification",
				toast.Info,
				3*time.Second,
			))

		case key.Matches(msg, m.keys.Warning):
			commands = append(commands, m.toastManager.Push(
				"Warning: This is a warning notification",
				toast.Warning,
				4*time.Second,
			))

		case key.Matches(msg, m.keys.Error):
			commands = append(commands, m.toastManager.Push(
				"Error: Something went wrong!",
				toast.Error,
				5*time.Second,
			))
		case key.Matches(msg, m.keys.Party):
			commands = append(commands, m.toastManager.Push(
				"It's time to PARTY! 🎊🎊🎊",
				PartyType,
				5*time.Second,
			))
		case key.Matches(msg, m.keys.DismissAll):
			commands = append(commands, m.toastManager.DismissAll())
		}

	case tea.WindowSizeMsg:
		m.toastManager.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.toastManager, cmd = m.toastManager.Update(msg)
	commands = append(commands, cmd)

	return m, tea.Batch(commands...)
}

func (m model) View() string {
	contentView := mainContentStyle.Render("Press keys to show different toasts:\n\ni: Info  w: Warning  e: Error  p: Party  space: Dismiss all  q: Quit")
	toastsView := m.toastManager.View()

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		contentView,
		"\n",
		toastsView,
	)

	return view
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
