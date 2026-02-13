package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
		m.bwSelectList.SetSize(msg.Width, msg.Height-2)
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 8
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case spinner.TickMsg:
		if m.tokenLoading || m.bwUnlocking {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case bwStatusMsg:
		m.bwChecking = false
		m.bwInstalled = msg.installed
		m.bwStatus = msg.status
		if msg.session != "" {
			m.bwSession = msg.session
		}
		if !msg.installed {
			m.statusMsg = "bw CLI not found"
			m.mode = bwLoginView
		} else if msg.status == "unauthenticated" {
			m.statusMsg = "Not logged in to Bitwarden"
			m.mode = bwLoginView
		} else if msg.status == "locked" {
			m.bwUnlocked = false
			m.statusMsg = "Vault locked"
			m.mode = bwPasswordView
			m.bwPwInput.Reset()
			m.bwPwInput.Focus()
			return m, m.bwPwInput.Cursor.BlinkCmd()
		} else if msg.status == "unlocked" {
			m.bwUnlocked = true
			m.statusMsg = "Vault unlocked"
			cmds = append(cmds, fetchBWItems(m.bwSession, ""))
		}

	case bwUnlockResultMsg:
		if msg.err != nil {
			m.bwUnlocking = false
			m.bwPwInput.Reset()
			m.bwUnlockErr = msg.err.Error()
			m.bwPwInput.Focus()
			return m, m.bwPwInput.Cursor.BlinkCmd()
		}
		// Keep bwUnlocking true to show spinner while fetching items
		m.bwSession = msg.session
		m.bwUnlocked = true
		m.bwStatus = "unlocked"
		m.bwUnlockErr = ""
		m.statusMsg = "Vault unlocked"
		return m, fetchBWItems(m.bwSession, "")

	case bwItemsFetchedMsg:
		m.bwUnlocking = false
		m.bwPwInput.Reset()
		if msg.err == nil {
			m.bwItems = msg.items
			m.updateBWSelectList()
		}
		if m.pendingAction != "" {
			action := m.pendingAction
			m.pendingAction = ""
			switch action {
			case "add":
				m.editingIndex = -1
				m.formClient = &Client{}
				m.updateBWSelectList()
				m.bwSelectList.ResetFilter()
				m.mode = bwSelectView
			case "edit":
				if item, ok := m.list.SelectedItem().(Client); ok {
					for i, c := range m.clients {
						if c.Name == item.Name && c.BitwardenItemID == item.BitwardenItemID {
							m.editingIndex = i
							break
						}
					}
					clientCopy := item
					m.formClient = &clientCopy
					m.form = buildClientForm(m.formClient)
					m.mode = formView
					return m, m.form.Init()
				}
			case "token":
				if item, ok := m.list.SelectedItem().(Client); ok {
					m.mode = tokenView
					m.tokenLoading = true
					m.tokenResult = nil
					return m, tea.Batch(m.spinner.Tick, requestToken(m.bwSession, item))
				}
			default:
				m.mode = listView
			}
		} else if m.mode == bwPasswordView || m.mode == bwLoginView {
			m.mode = listView
		}

	case tokenResponseMsg:
		m.tokenLoading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			if strings.HasPrefix(m.errorMsg, "bitwarden:") {
				m.bwUnlocked = false
			}
			m.prevMode = tokenView
			m.mode = errorView
			m.setErrorContent(m.errorMsg)
		} else {
			m.tokenResult = &msg.result
		}

	case clipboardCopyMsg:
		if msg.success {
			m.statusMsg = "Copied " + msg.what + " to clipboard"
		} else if msg.err != nil {
			m.statusMsg = "Failed to copy: " + msg.err.Error()
		}

	case clientsSavedMsg:
		if msg.err != nil {
			m.statusMsg = "Error saving: " + msg.err.Error()
		} else {
			m.statusMsg = "Client saved"
			m.updateList()
		}
	}

	switch m.mode {
	case listView:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	case bwSelectView:
		var cmd tea.Cmd
		m.bwSelectList, cmd = m.bwSelectList.Update(msg)
		cmds = append(cmds, cmd)
	case formView:
		if m.form != nil {
			form, cmd := m.form.Update(msg)
			if f, ok := form.(*huh.Form); ok {
				m.form = f
			}
			if m.form.State == huh.StateCompleted {
				return m.saveFormClient()
			}
			if m.form.State == huh.StateAborted {
				m.mode = listView
				return m, nil
			}
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case bwPasswordView:
		return m.handleBWPasswordKey(msg)
	case bwLoginView:
		return m.handleBWLoginKey(msg)
	case bwSelectView:
		return m.handleBWSelectKey(msg)
	case formView:
		return m.handleFormKey(msg)
	case tokenView:
		return m.handleTokenKey(msg)
	case errorView:
		return m.handleErrorKey(msg)
	case deleteView:
		return m.handleDeleteKey(msg)
	default:
		return m.handleListKey(msg)
	}
}

func (m model) handleBWPasswordKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.bwUnlocking {
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.mode = listView
		return m, nil
	case "enter":
		pw := m.bwPwInput.Value()
		if pw == "" {
			return m, nil
		}
		m.bwUnlocking = true
		m.bwUnlockErr = ""
		return m, tea.Batch(m.spinner.Tick, unlockBWVault(pw))
	}

	var cmd tea.Cmd
	m.bwPwInput, cmd = m.bwPwInput.Update(msg)
	return m, cmd
}

func (m model) handleBWLoginKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "r":
		return m, checkBWStatus(m.bwSession)
	case "esc":
		m.mode = listView
	}
	return m, nil
}

func (m model) handleBWSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Let the list handle filtering first
	if m.bwSelectList.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.bwSelectList, cmd = m.bwSelectList.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.mode = listView
		return m, nil
	case "enter":
		if item, ok := m.bwSelectList.SelectedItem().(BWItem); ok {
			m.formClient.BitwardenItemID = item.ID
			if m.formClient.Name == "" {
				m.formClient.Name = item.Name
			}
			if m.formClient.Issuer == "" && len(item.Login.URIs) > 0 {
				m.formClient.Issuer = item.Login.URIs[0].URI
			}
			m.form = buildClientForm(m.formClient)
			m.mode = formView
			return m, m.form.Init()
		}
	}

	var cmd tea.Cmd
	m.bwSelectList, cmd = m.bwSelectList.Update(msg)
	return m, cmd
}

func (m model) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.mode = listView
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		return m.saveFormClient()
	}
	if m.form.State == huh.StateAborted {
		m.mode = listView
		return m, nil
	}

	return m, cmd
}

func (m model) handleTokenKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = listView
		m.tokenResult = nil
		m.tokenLoading = false
		return m, nil
	case "c":
		if m.tokenResult != nil {
			return m, copyToClipboard(m.tokenResult.Token.AccessToken, "token")
		}
	case "h":
		if m.tokenResult != nil {
			header := "Authorization: Bearer " + m.tokenResult.Token.AccessToken
			return m, copyToClipboard(header, "header")
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleErrorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "enter":
		m.errorMsg = ""
		m.mode = m.prevMode
		if m.mode == tokenView {
			m.mode = listView
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m model) handleDeleteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "enter":
		if m.deleteIndex >= 0 && m.deleteIndex < len(m.clients) {
			m.clients = append(m.clients[:m.deleteIndex], m.clients[m.deleteIndex+1:]...)
			m.updateList()
			m.mode = listView
			m.statusMsg = "Client deleted"
			return m, saveClientsCmd(m.clients)
		}
		m.mode = listView
	case "n", "esc":
		m.mode = listView
	}
	return m, nil
}

func (m model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "enter":
		if item, ok := m.list.SelectedItem().(Client); ok {
			if !m.bwUnlocked {
				m.pendingAction = "token"
				return m.requireBWUnlock()
			}
			m.mode = tokenView
			m.tokenLoading = true
			m.tokenResult = nil
			return m, tea.Batch(m.spinner.Tick, requestToken(m.bwSession, item))
		}

	case "a":
		if !m.bwUnlocked {
			m.pendingAction = "add"
			return m.requireBWUnlock()
		}
		m.editingIndex = -1
		m.formClient = &Client{}
		m.updateBWSelectList()
		m.bwSelectList.ResetFilter()
		m.mode = bwSelectView
		return m, nil

	case "e":
		if item, ok := m.list.SelectedItem().(Client); ok {
			if !m.bwUnlocked {
				m.pendingAction = "edit"
				return m.requireBWUnlock()
			}
			for i, c := range m.clients {
				if c.Name == item.Name && c.BitwardenItemID == item.BitwardenItemID {
					m.editingIndex = i
					break
				}
			}
			clientCopy := item
			m.formClient = &clientCopy
			m.form = buildClientForm(m.formClient)
			m.mode = formView
			return m, m.form.Init()
		}

	case "d", "x":
		if item, ok := m.list.SelectedItem().(Client); ok {
			for i, c := range m.clients {
				if c.Name == item.Name && c.BitwardenItemID == item.BitwardenItemID {
					m.deleteIndex = i
					break
				}
			}
			m.mode = deleteView
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) requireBWUnlock() (tea.Model, tea.Cmd) {
	if m.bwChecking {
		m.statusMsg = "Checking Bitwarden status..."
		return m, nil
	}
	if m.bwStatus == "unauthenticated" || !m.bwInstalled {
		m.mode = bwLoginView
		return m, nil
	}
	m.mode = bwPasswordView
	m.bwPwInput.Reset()
	m.bwPwInput.Focus()
	m.bwUnlockErr = ""
	return m, m.bwPwInput.Cursor.BlinkCmd()
}

func (m model) saveFormClient() (tea.Model, tea.Cmd) {
	client := *m.formClient
	if m.editingIndex >= 0 && m.editingIndex < len(m.clients) {
		m.clients[m.editingIndex] = client
	} else {
		m.clients = append(m.clients, client)
	}
	m.updateList()
	m.mode = listView
	return m, saveClientsCmd(m.clients)
}
