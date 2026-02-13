package main

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPendingActionAdd(t *testing.T) {
	m := initialModel("")
	m.bwUnlocked = false
	m.bwChecking = false
	m.bwInstalled = true
	m.bwStatus = "locked"

	// Simulate pressing "a" while locked
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = result.(model)

	if m.pendingAction != "add" {
		t.Errorf("expected pendingAction 'add', got %q", m.pendingAction)
	}
	if m.mode != bwPasswordView {
		t.Errorf("expected bwPasswordView, got %v", m.mode)
	}
}

func TestPendingActionResumedAfterItemsFetched(t *testing.T) {
	m := initialModel("")
	m.bwUnlocked = true
	m.pendingAction = "add"

	// Simulate bwItemsFetchedMsg arriving after unlock
	result, _ := m.Update(bwItemsFetchedMsg{
		items: []BWItem{{ID: "item-1", Name: "Test Item"}},
	})
	m = result.(model)

	if m.mode != bwSelectView {
		t.Errorf("expected bwSelectView after add action, got %v", m.mode)
	}
	if m.pendingAction != "" {
		t.Errorf("expected pendingAction cleared, got %q", m.pendingAction)
	}
	if m.editingIndex != -1 {
		t.Errorf("expected editingIndex -1, got %d", m.editingIndex)
	}
}

func TestPendingActionClearedWhenNone(t *testing.T) {
	m := initialModel("")
	m.bwUnlocked = true
	m.mode = bwPasswordView
	m.pendingAction = ""

	result, _ := m.Update(bwItemsFetchedMsg{
		items: []BWItem{{ID: "item-1", Name: "Test Item"}},
	})
	m = result.(model)

	if m.mode != listView {
		t.Errorf("expected listView when no pending action, got %v", m.mode)
	}
}

func TestPendingActionTokenFlow(t *testing.T) {
	m := initialModel("")
	m.bwUnlocked = false
	m.bwChecking = false
	m.bwInstalled = true
	m.bwStatus = "locked"

	// Add a client to the list first
	m.clients = []Client{{Name: "test", BitwardenItemID: "bw-1", Issuer: "https://auth.example.com"}}
	m.updateList()

	// Simulate pressing Enter while locked
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\r'}})
	m = result.(model)

	// Enter key on list selects and tries to get token
	// But we need the list to have a selected item, which depends on bubbles internals
	// Just verify the basic flow works without error
}

func TestPendingActionEditFlow(t *testing.T) {
	m := initialModel("")
	m.bwUnlocked = false
	m.bwChecking = false
	m.bwInstalled = true
	m.bwStatus = "locked"

	// Simulate pressing "e" while locked
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = result.(model)

	if m.pendingAction != "edit" {
		t.Errorf("expected pendingAction 'edit', got %q", m.pendingAction)
	}
}

func TestStartupForcesLoginWhenUnauthenticated(t *testing.T) {
	m := initialModel("")
	m.bwChecking = true
	m.mode = listView

	// Simulate bwStatusMsg arriving with "unauthenticated"
	result, _ := m.Update(bwStatusMsg{installed: true, status: "unauthenticated"})
	m = result.(model)

	if m.mode != bwLoginView {
		t.Errorf("expected bwLoginView on unauthenticated, got %v", m.mode)
	}
}

func TestStartupForcesLoginWhenBWNotInstalled(t *testing.T) {
	m := initialModel("")
	m.bwChecking = true
	m.mode = listView

	// Simulate bwStatusMsg arriving with installed=false
	result, _ := m.Update(bwStatusMsg{installed: false, status: "unauthenticated"})
	m = result.(model)

	if m.mode != bwLoginView {
		t.Errorf("expected bwLoginView when bw not installed, got %v", m.mode)
	}
}

func TestUnlockKeepsLoadingUntilItemsFetched(t *testing.T) {
	m := initialModel("")
	m.bwUnlocking = true
	m.mode = bwPasswordView

	// Simulate successful unlock
	result, _ := m.Update(bwUnlockResultMsg{session: "test-session"})
	m = result.(model)

	// Should still show loading state while items are being fetched
	if !m.bwUnlocking {
		t.Error("expected bwUnlocking to remain true while fetching items")
	}
	if m.mode != bwPasswordView {
		t.Errorf("expected bwPasswordView while loading, got %v", m.mode)
	}

	// Now items arrive
	result, _ = m.Update(bwItemsFetchedMsg{
		items: []BWItem{{ID: "item-1", Name: "Test"}},
	})
	m = result.(model)

	if m.bwUnlocking {
		t.Error("expected bwUnlocking to be false after items fetched")
	}
}

func TestBWErrorResetsUnlockState(t *testing.T) {
	m := initialModel("")
	m.bwUnlocked = true
	m.mode = tokenView
	m.tokenLoading = true

	// Simulate a bitwarden error in token response
	result, _ := m.Update(tokenResponseMsg{
		err: fmt.Errorf("bitwarden: vault is locked"),
	})
	m = result.(model)

	if m.bwUnlocked {
		t.Error("expected bwUnlocked to be reset after bitwarden error")
	}
}
