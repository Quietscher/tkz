package main

import "time"

// viewMode represents the current TUI screen
type viewMode int

const (
	listView       viewMode = iota // Home screen with client list
	bwSelectView                   // Pick a Bitwarden item (bubbles/list)
	formView                       // Add/edit client form (huh)
	tokenView                      // Token request in progress / result display
	errorView                      // Error details
	deleteView                     // Confirm client deletion
	bwPasswordView                 // Master password prompt for locked vault
	bwLoginView                    // Instructions to run bw login (unauthenticated)
)

// Client represents a configured OAuth client (stored in clients.json)
type Client struct {
	Name              string `json:"name"`
	BitwardenItemID   string `json:"bitwarden_item_id"`
	Issuer            string `json:"issuer"`
	Scopes            string `json:"scopes"`
	ClientID          string `json:"client_id,omitempty"`
	ClientIDField     string `json:"client_id_field,omitempty"`
	ClientSecretField string `json:"client_secret_field,omitempty"`
}

// Title implements list.Item
func (c Client) Title() string { return c.Name }

// Description implements list.Item
func (c Client) Description() string {
	desc := c.Issuer
	if len(desc) > 50 {
		desc = desc[:47] + "..."
	}
	if c.ClientID != "" {
		desc += " (ID: " + c.ClientID + ")"
	}
	return desc
}

// FilterValue implements list.Item
func (c Client) FilterValue() string { return c.Name }

// BWItem represents a Bitwarden vault item (for search/select in form)
type BWItem struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Login BWLogin `json:"login"`
}

// Title implements list.Item
func (b BWItem) Title() string { return b.Name }

// Description implements list.Item
func (b BWItem) Description() string {
	if b.Login.Username != "" {
		return b.Login.Username
	}
	return b.ID
}

// FilterValue implements list.Item
func (b BWItem) FilterValue() string { return b.Name }

// BWLogin represents the login section of a Bitwarden item
type BWLogin struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	URIs     []BWURI `json:"uris"`
}

// BWURI represents a URI entry in a Bitwarden login item
type BWURI struct {
	URI string `json:"uri"`
}

// BWField represents a custom field in a Bitwarden item
type BWField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}

// BWFullItem represents a complete Bitwarden item with custom fields and notes
type BWFullItem struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Login  *BWLogin  `json:"login"`
	Fields []BWField `json:"fields"`
	Notes  string    `json:"notes"`
}

// BWCredentials holds credentials fetched from Bitwarden
type BWCredentials struct {
	ClientID     string
	ClientSecret string
	URIs         []string
}

// OIDCConfig represents relevant fields from an OpenID Connect discovery document
type OIDCConfig struct {
	TokenEndpoint string `json:"token_endpoint"`
	Issuer        string `json:"issuer"`
}

// TokenResponse represents an OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// TokenResult holds the complete result of a token request flow
type TokenResult struct {
	Token     TokenResponse
	Client    Client
	FetchedAt time.Time
}

// --- Bubble Tea message types ---

type bwStatusMsg struct {
	installed bool
	status    string // "unlocked", "locked", "unauthenticated"
	session   string
}

type bwUnlockResultMsg struct {
	session string
	err     error
}

type bwItemsFetchedMsg struct {
	items []BWItem
	err   error
}

type bwCredentialsFetchedMsg struct {
	creds BWCredentials
	err   error
}

type oidcDiscoveryMsg struct {
	config OIDCConfig
	err    error
}

type tokenResponseMsg struct {
	result TokenResult
	err    error
}

type clipboardCopyMsg struct {
	success bool
	what    string
	err     error
}

type clientsSavedMsg struct {
	err error
}
