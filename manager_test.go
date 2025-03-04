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

	// Headroom for message processing
	time.Sleep(10 * time.Millisecond)

	if view := m.View(); view != "" {
		t.Error("Expected empty view after toast expiry")
	}
}

func TestManagerHandleExpiredMsg(t *testing.T) {
	m := NewManager()
	toast1 := new("1", "Test Message 1", Info, 1*time.Second)
	toast2 := new("2", "Test Message 2", Warning, 2*time.Second)
	toast3 := new("3", "Test Message 3", Error, 3*time.Second)

	// Add toasts to manager
	m, _ = m.Update(msgPushed{Toast: toast1})
	m, _ = m.Update(msgPushed{Toast: toast2})
	m, _ = m.Update(msgPushed{Toast: toast3})

	// Verify we have 3 toasts
	if len(m.toasts) != 3 {
		t.Errorf("Expected 3 toasts, got %d", len(m.toasts))
	}

	// Remove the second toast
	m, _ = m.Update(msgExpired{ID: "2"})

	// Verify we now have 2 toasts
	if len(m.toasts) != 2 {
		t.Errorf("Expected 2 toasts after expiry, got %d", len(m.toasts))
	}

	// Verify the correct toast was removed
	for _, toast := range m.toasts {
		if toast.ID == "2" {
			t.Error("Toast with ID 2 should have been removed but was found")
		}
	}

	// Check the remaining toasts are correct
	foundToast1 := false
	foundToast3 := false

	for _, toast := range m.toasts {
		if toast.ID == "1" {
			foundToast1 = true
		}
		if toast.ID == "3" {
			foundToast3 = true
		}
	}

	if !foundToast1 || !foundToast3 {
		t.Error("Expected to find toast1 and toast3, but one or both were missing")
	}
}
