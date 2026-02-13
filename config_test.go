package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadClientsEmpty(t *testing.T) {
	tmp := t.TempDir()
	clients, err := loadClientsFrom(filepath.Join(tmp, "clients.json"))
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(clients) != 0 {
		t.Fatalf("expected empty slice, got %d clients", len(clients))
	}
}

func TestSaveAndLoadClients(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "clients.json")

	clients := []Client{
		{
			Name:            "test-client",
			BitwardenItemID: "abc-123",
			Issuer:          "https://auth.example.com",
			Scopes:          "openid profile",
		},
		{
			Name:            "another-client",
			BitwardenItemID: "def-456",
			Issuer:          "https://other.example.com/realms/dev",
			Scopes:          "email",
		},
	}

	err := saveClientsTo(path, clients)
	if err != nil {
		t.Fatalf("saveClientsTo failed: %v", err)
	}

	loaded, err := loadClientsFrom(path)
	if err != nil {
		t.Fatalf("loadClientsFrom failed: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(loaded))
	}

	if loaded[0].Name != "test-client" {
		t.Errorf("expected name 'test-client', got '%s'", loaded[0].Name)
	}
	if loaded[0].BitwardenItemID != "abc-123" {
		t.Errorf("expected BW item ID 'abc-123', got '%s'", loaded[0].BitwardenItemID)
	}
	if loaded[0].Issuer != "https://auth.example.com" {
		t.Errorf("expected issuer 'https://auth.example.com', got '%s'", loaded[0].Issuer)
	}
	if loaded[0].Scopes != "openid profile" {
		t.Errorf("expected scopes 'openid profile', got '%s'", loaded[0].Scopes)
	}

	if loaded[1].Name != "another-client" {
		t.Errorf("expected name 'another-client', got '%s'", loaded[1].Name)
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "subdir", "clients.json")

	err := saveClientsTo(path, []Client{{Name: "test"}})
	if err != nil {
		t.Fatalf("saveClientsTo failed to create directory: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected file to exist after save")
	}
}

func TestLoadCorruptedFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "clients.json")

	err := os.WriteFile(path, []byte("not valid json"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = loadClientsFrom(path)
	if err == nil {
		t.Fatal("expected error for corrupted JSON")
	}
}

func TestSaveAndLoadClientsWithFieldMapping(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "clients.json")

	clients := []Client{
		{
			Name:              "custom-fields",
			BitwardenItemID:   "bw-123",
			Issuer:            "https://auth.example.com",
			Scopes:            "openid",
			ClientID:          "manual-id",
			ClientIDField:     "fields.my_client_id",
			ClientSecretField: "fields.api_key",
		},
		{
			Name:            "defaults",
			BitwardenItemID: "bw-456",
			Issuer:          "https://other.example.com",
			Scopes:          "profile",
		},
	}

	if err := saveClientsTo(path, clients); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := loadClientsFrom(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded[0].ClientID != "manual-id" {
		t.Errorf("expected ClientID 'manual-id', got '%s'", loaded[0].ClientID)
	}
	if loaded[0].ClientIDField != "fields.my_client_id" {
		t.Errorf("expected ClientIDField 'fields.my_client_id', got '%s'", loaded[0].ClientIDField)
	}
	if loaded[0].ClientSecretField != "fields.api_key" {
		t.Errorf("expected ClientSecretField 'fields.api_key', got '%s'", loaded[0].ClientSecretField)
	}

	if loaded[1].ClientID != "" {
		t.Errorf("expected empty ClientID, got '%s'", loaded[1].ClientID)
	}
	if loaded[1].ClientIDField != "" {
		t.Errorf("expected empty ClientIDField, got '%s'", loaded[1].ClientIDField)
	}
}

func TestLoadLegacyClientsFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "clients.json")
	legacyJSON := `{
		"clients": [
			{
				"name": "old-client",
				"bitwarden_item_id": "bw-old",
				"issuer": "https://legacy.example.com",
				"scopes": "openid"
			}
		]
	}`
	if err := os.WriteFile(path, []byte(legacyJSON), 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := loadClientsFrom(path)
	if err != nil {
		t.Fatalf("load legacy failed: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 client, got %d", len(loaded))
	}
	if loaded[0].ClientID != "" {
		t.Errorf("expected empty ClientID for legacy client, got '%s'", loaded[0].ClientID)
	}
	if loaded[0].ClientIDField != "" {
		t.Errorf("expected empty ClientIDField, got '%s'", loaded[0].ClientIDField)
	}
	if loaded[0].ClientSecretField != "" {
		t.Errorf("expected empty ClientSecretField, got '%s'", loaded[0].ClientSecretField)
	}
}

func TestSaveOverwritesExisting(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "clients.json")

	err := saveClientsTo(path, []Client{{Name: "first"}, {Name: "second"}})
	if err != nil {
		t.Fatal(err)
	}

	err = saveClientsTo(path, []Client{{Name: "only"}})
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := loadClientsFrom(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 {
		t.Fatalf("expected 1 client after overwrite, got %d", len(loaded))
	}
	if loaded[0].Name != "only" {
		t.Errorf("expected 'only', got '%s'", loaded[0].Name)
	}
}
