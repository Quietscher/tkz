package main

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	accentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	tokenStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	labelStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).PaddingLeft(1).PaddingRight(1)
	spinnerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	tokenBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(1, 2)
)
