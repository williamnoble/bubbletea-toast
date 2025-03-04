package toast

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type toast struct {
	ID       string
	Type     Type
	Message  string
	Duration time.Duration
	timer    timer.Model
	visible  bool
}

func new(id string, message string, toastType Type, duration time.Duration) toast {
	return toast{
		ID:       id,
		Type:     toastType,
		Message:  message,
		Duration: duration,
		timer:    timer.NewWithInterval(duration, time.Millisecond*100),
		visible:  true,
	}
}

func (t toast) update(msg tea.Msg) (toast, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		t.timer, cmd = t.timer.Update(msg)
		return t, cmd
	case timer.TimeoutMsg:
		t.visible = false
		return t, func() tea.Msg { return msgExpired{ID: t.ID} }
	}

	return t, nil
}

func (t toast) view(customStyle *CustomToastStyle) string {
	if !t.visible {
		return ""
	}

	prefix, style := toastPrefixAndStyle(t.Type, customStyle)
	content := fmt.Sprintf("%s %s", prefix, t.Message)

	return style.Render(content)
}

func (t toast) init() tea.Cmd {
	return t.timer.Init()
}

func toastPrefixAndStyle(toastType Type, customStyle *CustomToastStyle) (string, lipgloss.Style) {
	var style lipgloss.Style
	var prefix string

	baseStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		Width(30)
	infoStyle := baseStyle.
		Foreground(lipgloss.Color("12")).
		BorderForeground(lipgloss.Color("12"))
	warningStyle := baseStyle.
		Foreground(lipgloss.Color("11")).
		BorderForeground(lipgloss.Color("11"))
	errorStyle := baseStyle.
		Foreground(lipgloss.Color("9")).
		BorderForeground(lipgloss.Color("9"))

	if customStyle != nil {
		prefix = customStyle.Prefix
		style = customStyle.Style
		return prefix, style
	}

	switch toastType {
	case Info:
		style = infoStyle
		prefix = "ℹ"
	case Warning:
		style = warningStyle
		prefix = "⚠"
	case Error:
		style = errorStyle
		prefix = "✗"
	default:
		style = baseStyle
		prefix = ""
	}

	return prefix, style
}
