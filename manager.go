package toast

import (
	"fmt"
	"github.com/charmbracelet/bubbles/timer"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type position int

const (
	TopLeft position = iota
	TopCenter
	TopRight
	BottomLeft
	BottomCenter
	BottomRight
)

type msgPushed struct {
	Toast toast
}

type msgDismissAll struct{}

type Type int

type CustomToastStyle struct {
	Prefix string
	Style  lipgloss.Style
}

const (
	Info Type = iota
	Warning
	Error
)

// Manager handles the lifecycle of toasts
type Manager struct {
	toasts    []toast
	maxToasts int
	counter   int64

	position position
	width    int
	height   int

	customStyles map[Type]CustomToastStyle
}

// NewManager creates a new toast manager
func NewManager() Manager {
	return Manager{
		toasts: []toast{},

		maxToasts: 3,
		position:  TopRight,
		width:     30,

		customStyles: make(map[Type]CustomToastStyle),
	}
}

// WithPosition sets the position for toast notifications
func (m Manager) WithPosition(pos position) Manager {
	m.position = pos
	return m
}

// WithMaxToasts sets the maximum number of visible toasts
func (m Manager) WithMaxToasts(max int) Manager {
	m.maxToasts = max
	return m
}

// WithWidth sets the width of toast notifications
func (m Manager) WithWidth(width int) Manager {
	m.width = width
	return m
}

func (m Manager) WithStyle(prefixType Type, prefix string, style lipgloss.Style) Manager {
	m.customStyles[prefixType] = CustomToastStyle{
		Prefix: prefix,
		Style:  style,
	}
	return m
}

func (m Manager) DismissAll() tea.Cmd {
	return func() tea.Msg {
		return msgDismissAll{}
	}
}

// SetSize updates the width and height for toast positioning
func (m Manager) SetSize(width, height int) Manager {
	m.width = width
	m.height = height
	return m
}

// Push adds a new toast to the queue
func (m Manager) Push(message string, toastType Type, duration time.Duration) tea.Cmd {
	m.counter++
	id := fmt.Sprintf("toast-%d", m.counter)
	toast := new(id, message, toastType, duration)
	return func() tea.Msg {
		return msgPushed{Toast: toast}
	}
}

func (m Manager) Update(msg tea.Msg) (Manager, tea.Cmd) {
	var commands []tea.Cmd

	switch msg := msg.(type) {

	case msgPushed:
		m.toasts = append(m.toasts, msg.Toast)
		commands = append(commands, msg.Toast.init())

	case timer.TickMsg, timer.TimeoutMsg:
		for i, t := range m.toasts {
			var cmd tea.Cmd
			m.toasts[i], cmd = t.update(msg)
			if cmd != nil {
				commands = append(commands, cmd)
			}
		}

	case msgExpired:
		for i, t := range m.toasts {
			if t.ID == msg.ID {
				m.toasts = append(m.toasts[:i], m.toasts[i+1:]...)
				break
			}
		}

	case msgDismissAll:
		m.toasts = []toast{}
	}

	return m, tea.Batch(commands...)
}

func (m Manager) View() string {
	visibleCount := 0
	var visibleToasts []string

	for _, t := range m.toasts {
		if t.visible {
			visibleCount++
			var customToastStyle *CustomToastStyle
			if style, exists := m.customStyles[t.Type]; exists {
				customToastStyle = &style
			}

			if visibleCount <= m.maxToasts {
				visibleToasts = append(visibleToasts, t.view(customToastStyle))
			} else {
				break
			}
		}
	}

	if len(visibleToasts) == 0 {
		return ""
	}

	joined := strings.Join(visibleToasts, "\n")

	var view string
	var alignment lipgloss.Position
	switch m.position {
	case TopLeft, BottomLeft:
		alignment = lipgloss.Left
	case TopCenter, BottomCenter:
		alignment = lipgloss.Center
	case TopRight, BottomRight:
		alignment = lipgloss.Right
	}

	view = lipgloss.NewStyle().Align(alignment).Render(joined)
	return view
}
