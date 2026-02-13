package main

import (
	"testing"
)

func TestParseBWStatusDetail(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		wantStatus string
		wantErr    bool
	}{
		{
			name:       "unlocked vault",
			json:       `{"serverUrl":"https://vault.bitwarden.com","lastSync":"2026-02-12T10:00:00.000Z","userEmail":"user@example.com","userId":"abc-123","status":"unlocked"}`,
			wantStatus: "unlocked",
		},
		{
			name:       "locked vault",
			json:       `{"serverUrl":"https://vault.bitwarden.com","lastSync":"2026-02-12T10:00:00.000Z","userEmail":"user@example.com","userId":"abc-123","status":"locked"}`,
			wantStatus: "locked",
		},
		{
			name:       "unauthenticated",
			json:       `{"serverUrl":"https://vault.bitwarden.com","lastSync":null,"userEmail":null,"userId":null,"status":"unauthenticated"}`,
			wantStatus: "unauthenticated",
		},
		{
			name:    "invalid json",
			json:    `not json`,
			wantErr: true,
		},
		{
			name:       "empty status",
			json:       `{"status":""}`,
			wantStatus: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := parseBWStatusDetail([]byte(tt.json))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if status != tt.wantStatus {
				t.Errorf("expected status=%q, got %q", tt.wantStatus, status)
			}
		})
	}
}

func TestParseBWItems(t *testing.T) {
	json := `[
		{
			"id": "item-1",
			"name": "Keycloak Dev",
			"type": 1,
			"login": {
				"username": "my-client-id",
				"password": "my-client-secret",
				"uris": [{"uri": "https://auth.example.com", "match": null}]
			}
		},
		{
			"id": "item-2",
			"name": "Auth0 Staging",
			"type": 1,
			"login": {
				"username": "auth0-id",
				"password": "auth0-secret",
				"uris": []
			}
		}
	]`

	items, err := parseBWItems([]byte(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].ID != "item-1" {
		t.Errorf("expected ID 'item-1', got '%s'", items[0].ID)
	}
	if items[0].Name != "Keycloak Dev" {
		t.Errorf("expected name 'Keycloak Dev', got '%s'", items[0].Name)
	}
	if items[0].Login.Username != "my-client-id" {
		t.Errorf("expected username 'my-client-id', got '%s'", items[0].Login.Username)
	}
	if len(items[0].Login.URIs) != 1 || items[0].Login.URIs[0].URI != "https://auth.example.com" {
		t.Errorf("expected URI 'https://auth.example.com', got %v", items[0].Login.URIs)
	}

	if items[1].Name != "Auth0 Staging" {
		t.Errorf("expected name 'Auth0 Staging', got '%s'", items[1].Name)
	}
}

func TestParseBWItemsEmpty(t *testing.T) {
	items, err := parseBWItems([]byte(`[]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestParseBWItemsInvalid(t *testing.T) {
	_, err := parseBWItems([]byte(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseBWItem(t *testing.T) {
	json := `{
		"id": "item-1",
		"name": "Keycloak Dev",
		"type": 1,
		"login": {
			"username": "my-client-id",
			"password": "my-client-secret",
			"uris": [
				{"uri": "https://auth.example.com/realms/dev", "match": null},
				{"uri": "https://auth.example.com/admin", "match": null}
			]
		}
	}`

	creds, err := parseBWItemCredentials([]byte(json))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.ClientID != "my-client-id" {
		t.Errorf("expected ClientID 'my-client-id', got '%s'", creds.ClientID)
	}
	if creds.ClientSecret != "my-client-secret" {
		t.Errorf("expected ClientSecret 'my-client-secret', got '%s'", creds.ClientSecret)
	}
	if len(creds.URIs) != 2 {
		t.Fatalf("expected 2 URIs, got %d", len(creds.URIs))
	}
	if creds.URIs[0] != "https://auth.example.com/realms/dev" {
		t.Errorf("expected first URI 'https://auth.example.com/realms/dev', got '%s'", creds.URIs[0])
	}
}

func TestParseBWItemNoLogin(t *testing.T) {
	json := `{
		"id": "item-1",
		"name": "Secure Note",
		"type": 2,
		"login": null
	}`

	_, err := parseBWItemCredentials([]byte(json))
	if err == nil {
		t.Fatal("expected error for item without login")
	}
}

func TestParseBWFullItem(t *testing.T) {
	t.Run("full item with login, fields, notes", func(t *testing.T) {
		data := []byte(`{
			"id": "item-1",
			"name": "Keycloak Dev",
			"login": {
				"username": "my-client",
				"password": "my-secret",
				"uris": [{"uri": "https://auth.example.com"}]
			},
			"fields": [
				{"name": "client_id", "value": "custom-id", "type": 0},
				{"name": "api_key", "value": "hidden-key", "type": 1}
			],
			"notes": "some notes"
		}`)
		item, err := parseBWFullItem(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.ID != "item-1" {
			t.Errorf("expected ID 'item-1', got '%s'", item.ID)
		}
		if item.Login == nil || item.Login.Username != "my-client" {
			t.Error("expected login.username = 'my-client'")
		}
		if len(item.Fields) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(item.Fields))
		}
		if item.Fields[0].Name != "client_id" || item.Fields[0].Value != "custom-id" {
			t.Error("expected first field name=client_id, value=custom-id")
		}
		if item.Notes != "some notes" {
			t.Errorf("expected notes 'some notes', got '%s'", item.Notes)
		}
	})

	t.Run("no custom fields", func(t *testing.T) {
		data := []byte(`{"id": "item-2", "name": "Simple", "login": {"username": "u", "password": "p"}}`)
		item, err := parseBWFullItem(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(item.Fields) != 0 {
			t.Errorf("expected 0 fields, got %d", len(item.Fields))
		}
	})

	t.Run("no login section", func(t *testing.T) {
		data := []byte(`{"id": "item-3", "name": "Note", "login": null, "notes": "just a note"}`)
		item, err := parseBWFullItem(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.Login != nil {
			t.Error("expected nil login")
		}
		if item.Notes != "just a note" {
			t.Errorf("expected notes 'just a note', got '%s'", item.Notes)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := parseBWFullItem([]byte(`not json`))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestResolveBWField(t *testing.T) {
	fullItem := &BWFullItem{
		Login: &BWLogin{
			Username: "the-username",
			Password: "the-password",
			URIs:     []BWURI{{URI: "https://example.com"}},
		},
		Fields: []BWField{
			{Name: "client_id", Value: "custom-client-id", Type: 0},
			{Name: "api_key", Value: "secret-api-key", Type: 1},
		},
		Notes: "some notes content",
	}

	tests := []struct {
		name      string
		item      *BWFullItem
		fieldPath string
		want      string
		wantErr   bool
	}{
		{"login.username", fullItem, "login.username", "the-username", false},
		{"login.password", fullItem, "login.password", "the-password", false},
		{"custom field text", fullItem, "fields.client_id", "custom-client-id", false},
		{"custom field hidden", fullItem, "fields.api_key", "secret-api-key", false},
		{"notes", fullItem, "notes", "some notes content", false},
		{"empty path", fullItem, "", "", false},
		{"missing custom field", fullItem, "fields.nonexistent", "", true},
		{"unsupported path", fullItem, "foo.bar", "", true},
		{"login.username nil login", &BWFullItem{}, "login.username", "", true},
		{"login.password nil login", &BWFullItem{}, "login.password", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveBWField(tt.item, tt.fieldPath)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
