package main

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

func checkBWStatus(session string) tea.Cmd {
	return func() tea.Msg {
		installed := CheckBWInstalled()
		if !installed {
			return bwStatusMsg{installed: false, status: "unauthenticated"}
		}
		status := CheckBWStatusDetail(session)
		return bwStatusMsg{installed: installed, status: status, session: session}
	}
}

func unlockBWVault(password string) tea.Cmd {
	return func() tea.Msg {
		session, err := UnlockBWVault(password)
		return bwUnlockResultMsg{session: session, err: err}
	}
}

func fetchBWItems(session string, search string) tea.Cmd {
	return func() tea.Msg {
		items, err := FetchBWItems(session, search)
		return bwItemsFetchedMsg{items: items, err: err}
	}
}

func requestToken(session string, client Client) tea.Cmd {
	return func() tea.Msg {
		raw, err := fetchBWRawItem(session, client.BitwardenItemID)
		if err != nil {
			return tokenResponseMsg{err: fmt.Errorf("bitwarden: %w", err)}
		}

		fullItem, err := parseBWFullItem(raw)
		if err != nil {
			return tokenResponseMsg{err: fmt.Errorf("bitwarden parse: %w", err)}
		}

		// Resolve client_id: manual override takes precedence
		clientID := client.ClientID
		if clientID == "" {
			fieldPath := client.ClientIDField
			if fieldPath == "" {
				fieldPath = "login.username"
			}
			clientID, err = ResolveBWField(fullItem, fieldPath)
			if err != nil {
				return tokenResponseMsg{err: fmt.Errorf("resolve client_id (%s): %w", fieldPath, err)}
			}
		}

		// Resolve client_secret: always from Bitwarden
		secretFieldPath := client.ClientSecretField
		if secretFieldPath == "" {
			secretFieldPath = "login.password"
		}
		clientSecret, err := ResolveBWField(fullItem, secretFieldPath)
		if err != nil {
			return tokenResponseMsg{err: fmt.Errorf("resolve client_secret (%s): %w", secretFieldPath, err)}
		}

		oidc, err := DiscoverOIDC(client.Issuer)
		if err != nil {
			return tokenResponseMsg{err: fmt.Errorf("oidc discovery: %w", err)}
		}

		token, err := RequestToken(oidc.TokenEndpoint, clientID, clientSecret, client.Scopes)
		if err != nil {
			return tokenResponseMsg{err: fmt.Errorf("token request: %w", err)}
		}

		return tokenResponseMsg{
			result: TokenResult{Token: *token, Client: client, FetchedAt: time.Now()},
		}
	}
}

func copyToClipboard(text string, what string) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(text)
		return clipboardCopyMsg{success: err == nil, what: what, err: err}
	}
}

func saveClientsCmd(clients []Client) tea.Cmd {
	return func() tea.Msg {
		err := saveClients(clients)
		return clientsSavedMsg{err: err}
	}
}
