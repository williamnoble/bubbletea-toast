package toast

import (
	"github.com/charmbracelet/bubbles/timer"
	"testing"
	"time"
)

func TestManagerPush(t *testing.T) {
	m := NewManager()
	msg := m.Push("Test message", Info, 1*time.Second)
	cmd := msg()
	if _, ok := cmd.(msgPushed); !ok {
		t.Errorf("Expected msgPushed, got %T", cmd)
	}
}

func TestManagerUpdate(t *testing.T) {
	m := NewManager()
	pushMsg := msgPushed{Toast: new("1", "Test", Info, 1*time.Second)}
	m, _ = m.Update(pushMsg)

	if len(m.toasts) != 1 {
		t.Errorf("Expected 1 toast, got %d", len(m.toasts))
	}

	expireMsg := msgExpired{ID: "1"}
	m, _ = m.Update(expireMsg)

	if len(m.toasts) != 0 {
		t.Errorf("Expected 0 toasts after expiry, got %d", len(m.toasts))
	}
}

func TestManagerView(t *testing.T) {
	m := NewManager()

	toast := new("1", "Test message", Info, 1*time.Second)
	m.toasts = append(m.toasts, toast)

	view := m.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}

	m.toasts[0].visible = false
	view = m.View()
	if view != "" {
		t.Error("Expected empty view with invisible toast")
	}
}

func TestManagerDismissAll(t *testing.T) {
	m := NewManager()

	pushCmd := m.Push("Test message", Info, 1*time.Second)
	m, _ = m.Update(pushCmd())

	if view := m.View(); view == "" {
		t.Error("Expected non-empty view after pushing toast")
	}

	dismissCmd := m.DismissAll()
	m, _ = m.Update(dismissCmd())

	if view := m.View(); view != "" {
		t.Error("Expected empty view after dismissing all toasts")
	}
}
func TestManagerExpiry(t *testing.T) {
	m := NewManager()

	// Add a toast with very short duration
	pushCmd := m.Push("Test message", Info, 10*time.Millisecond)
	m, _ = m.Update(pushCmd())

	if view := m.View(); view == "" {
		t.Error("Expected non-empty view after pushing toast")
	}

	time.Sleep(20 * time.Millisecond)

	timeoutMsg := timer.TimeoutMsg{}
	m, _ = m.Update(timeoutMsg)

	// Head room for message processing
	time.Sleep(10 * time.Millisecond)

	if view := m.View(); view != "" {
		t.Error("Expected empty view after toast expiry")
	}
}
