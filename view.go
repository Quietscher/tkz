package main

import (
	"fmt"
	"strings"
)

func (m model) View() string {
	switch m.mode {
	case bwPasswordView:
		return m.viewBWPassword()
	case bwLoginView:
		return m.viewBWLogin()
	case bwSelectView:
		return m.viewBWSelect()
	case formView:
		return m.viewForm()
	case tokenView:
		return m.viewToken()
	case errorView:
		return m.viewError()
	case deleteView:
		return m.viewDelete()
	default:
		return m.viewList()
	}
}

func (m model) viewBWPassword() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Unlock Bitwarden Vault"))
	b.WriteString("\n\n")

	if m.bwUnlocking {
		b.WriteString(m.spinner.View())
		if m.bwUnlocked {
			b.WriteString(" Loading vault items...")
		} else {
			b.WriteString(" Unlocking vault...")
		}
		b.WriteString("\n\n")
		return b.String()
	}

	if m.bwUnlockErr != "" {
		b.WriteString(errorStyle.Render(m.bwUnlockErr))
		b.WriteString("\n\n")
	}

	b.WriteString("Master password: ")
	b.WriteString(m.bwPwInput.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("enter: unlock • esc: back • ctrl+c: quit"))

	return b.String()
}

func (m model) viewBWLogin() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Bitwarden Not Logged In"))
	b.WriteString("\n\n")

	if !m.bwInstalled {
		b.WriteString("The Bitwarden CLI (")
		b.WriteString(accentStyle.Render("bw"))
		b.WriteString(") was not found.\n\n")
		b.WriteString("Install it:\n\n")
		b.WriteString(accentStyle.Render("  brew install bitwarden-cli"))
		b.WriteString("\n\n")
	} else {
		b.WriteString("You need to log in to Bitwarden first.\n\n")
		b.WriteString("Run in another terminal:\n\n")
		b.WriteString(accentStyle.Render("  bw login"))
		b.WriteString("\n\n")
		b.WriteString("Then press ")
		b.WriteString(accentStyle.Render("r"))
		b.WriteString(" to retry.")
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("r: retry • esc: back • q: quit"))
	return b.String()
}

func (m model) viewBWSelect() string {
	return m.bwSelectList.View()
}

func (m model) viewForm() string {
	var b strings.Builder
	title := "Add OAuth Client"
	if m.editingIndex >= 0 {
		title = "Edit OAuth Client"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")
	if m.form != nil {
		b.WriteString(m.form.View())
	}
	return b.String()
}

func (m model) viewToken() string {
	var b strings.Builder

	if m.tokenLoading {
		b.WriteString(titleStyle.Render("Requesting Token"))
		b.WriteString("\n\n")
		b.WriteString(m.spinner.View())
		b.WriteString(" Fetching token...")
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc: cancel"))
		return b.String()
	}

	if m.tokenResult != nil {
		b.WriteString(titleStyle.Render("Token for " + m.tokenResult.Client.Name))
		b.WriteString("\n\n")

		truncated := m.tokenResult.Token.AccessToken
		if len(truncated) > 80 {
			truncated = truncated[:40] + "..." + truncated[len(truncated)-40:]
		}

		content := fmt.Sprintf(
			"%s %s\n%s %d seconds\n%s %s",
			accentStyle.Render("Type:"),
			m.tokenResult.Token.TokenType,
			accentStyle.Render("Expires:"),
			m.tokenResult.Token.ExpiresIn,
			accentStyle.Render("Token:"),
			tokenStyle.Render(truncated),
		)

		if m.tokenResult.Token.Scope != "" {
			content += fmt.Sprintf("\n%s %s",
				accentStyle.Render("Scope:"),
				m.tokenResult.Token.Scope,
			)
		}

		b.WriteString(tokenBoxStyle.Render(content))
		b.WriteString("\n\n")

		if m.statusMsg != "" {
			b.WriteString(successStyle.Render(m.statusMsg))
			b.WriteString("\n\n")
		}

		b.WriteString(helpStyle.Render("c: copy token • h: copy as Authorization header • esc: back"))
	}

	return b.String()
}

func (m model) viewError() string {
	var b strings.Builder
	b.WriteString(errorStyle.Render("Error"))
	b.WriteString("\n\n")
	b.WriteString(m.viewport.View())
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("esc/enter: dismiss"))
	return b.String()
}

func (m model) viewDelete() string {
	var b strings.Builder
	if m.deleteIndex >= 0 && m.deleteIndex < len(m.clients) {
		client := m.clients[m.deleteIndex]
		b.WriteString(warningStyle.Render("Delete client: " + client.Name + "?"))
	} else {
		b.WriteString(warningStyle.Render("Delete client?"))
	}
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("y/enter: delete • n/esc: cancel"))
	return b.String()
}

func (m model) viewList() string {
	if m.bwChecking {
		var b strings.Builder
		b.WriteString(titleStyle.Render("tkz"))
		b.WriteString("\n\n")
		b.WriteString(m.spinner.View())
		b.WriteString(" Connecting to Bitwarden...")
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("ctrl+c: quit"))
		return b.String()
	}

	var b strings.Builder
	b.WriteString(m.list.View())
	b.WriteString("\n")

	if m.statusMsg != "" {
		b.WriteString(dimStyle.Render(m.statusMsg))
		b.WriteString(" ")
	}

	bwIndicator := ""
	if !m.bwInstalled {
		bwIndicator = errorStyle.Render("[bw not found]")
	} else if !m.bwUnlocked {
		bwIndicator = warningStyle.Render("[vault locked]")
	} else {
		bwIndicator = successStyle.Render("[vault unlocked]")
	}
	b.WriteString(bwIndicator)
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter: get token • a: add • e: edit • d: delete • /: filter • q: quit"))

	return b.String()
}
