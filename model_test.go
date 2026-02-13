package main

import (
	"strings"
	"testing"
)

func TestSetErrorContent(t *testing.T) {
	t.Run("wraps long text to viewport width", func(t *testing.T) {
		m := initialModel("")
		m.viewport.Width = 40

		longError := strings.Repeat("error ", 20) // 120 chars
		m.setErrorContent(longError)

		content := m.viewport.View()
		lines := strings.Split(content, "\n")
		if len(lines) < 2 {
			t.Errorf("expected wrapped text (multiple lines), got %d line(s)", len(lines))
		}
	})

	t.Run("uses default width when viewport uninitialized", func(t *testing.T) {
		m := initialModel("")
		m.viewport.Width = 0

		m.setErrorContent("some error text")

		content := m.viewport.View()
		if !strings.Contains(content, "some error text") {
			t.Error("expected error text in viewport content")
		}
	})

	t.Run("preserves line breaks in error", func(t *testing.T) {
		m := initialModel("")
		m.viewport.Width = 80
		m.viewport.Height = 20

		m.setErrorContent("line one\nline two\nline three")

		content := m.viewport.View()
		if !strings.Contains(content, "line one") {
			t.Error("expected 'line one' in content")
		}
		if !strings.Contains(content, "line two") {
			t.Error("expected 'line two' in content")
		}
	})
}
