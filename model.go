package main

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	list     list.Model
	spinner  spinner.Model
	viewport viewport.Model
	width    int
	height   int
	mode     viewMode
	prevMode viewMode

	clients   []Client
	statusMsg string
	errorMsg  string

	bwSession    string
	bwInstalled  bool
	bwUnlocked   bool
	bwChecking   bool
	bwStatus     string // "unlocked", "locked", "unauthenticated"
	bwItems      []BWItem
	bwSelectList list.Model
	bwPwInput    textinput.Model
	bwUnlocking  bool
	bwUnlockErr  string

	form          *huh.Form
	editingIndex  int
	formClient    *Client
	pendingAction string

	tokenResult  *TokenResult
	tokenLoading bool

	deleteIndex int
}

func initialModel(bwSession string) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	clients, _ := loadClients()

	delegate := list.NewDefaultDelegate()
	l := list.New(clientsToItems(clients), delegate, 0, 0)
	l.Title = "tkz"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = titleStyle

	vp := viewport.New(80, 20)

	pwInput := textinput.New()
	pwInput.Placeholder = "Master password"
	pwInput.EchoMode = textinput.EchoPassword
	pwInput.EchoCharacter = '*'
	pwInput.Width = 40

	bwList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	bwList.Title = "Select Bitwarden Item"
	bwList.SetShowStatusBar(true)
	bwList.SetFilteringEnabled(true)
	bwList.SetShowHelp(false)
	bwList.Styles.Title = titleStyle

	return model{
		list:         l,
		spinner:      s,
		viewport:     vp,
		bwPwInput:    pwInput,
		bwSelectList: bwList,
		bwChecking:   true,
		mode:         listView,
		clients:      clients,
		bwSession:    bwSession,
		editingIndex: -1,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		checkBWStatus(m.bwSession),
	)
}

func clientsToItems(clients []Client) []list.Item {
	items := make([]list.Item, len(clients))
	for i, c := range clients {
		items[i] = c
	}
	return items
}

func bwItemsToListItems(items []BWItem) []list.Item {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	return listItems
}

func (m *model) updateList() {
	m.list.SetItems(clientsToItems(m.clients))
}

func (m *model) updateBWSelectList() {
	m.bwSelectList.SetItems(bwItemsToListItems(m.bwItems))
}

func (m *model) setErrorContent(errMsg string) {
	width := m.viewport.Width
	if width <= 0 {
		width = 76
	}
	wrapped := lipgloss.NewStyle().Width(width).Render(errMsg)
	m.viewport.SetContent(wrapped)
	m.viewport.GotoTop()
}

func buildClientForm(client *Client) *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Display Name").
				Value(&client.Name).
				Placeholder("e.g., keycloak-dev"),

			huh.NewInput().
				Title("Client ID").
				Value(&client.ClientID).
				Placeholder("leave empty to pull from Bitwarden").
				Description("Manual override — skips BW lookup for client_id"),

			huh.NewInput().
				Title("Client ID Field").
				Value(&client.ClientIDField).
				Placeholder("login.username").
				Description("BW field: login.username, login.password, fields.<name>, notes"),

			huh.NewInput().
				Title("Client Secret Field").
				Value(&client.ClientSecretField).
				Placeholder("login.password").
				Description("BW field: login.username, login.password, fields.<name>, notes"),

			huh.NewInput().
				Title("Issuer URL").
				Value(&client.Issuer).
				Placeholder("https://auth.example.com/realms/myrealm"),

			huh.NewInput().
				Title("Scopes (space-separated)").
				Value(&client.Scopes).
				Placeholder("openid profile email"),
		),
	).WithTheme(huh.ThemeDracula()).WithWidth(60)
}
