package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type clientsFile struct {
	Clients []Client `json:"clients"`
}

func getConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "tkz")
}

func getClientsPath() string {
	return filepath.Join(getConfigDir(), "clients.json")
}

func loadClients() ([]Client, error) {
	return loadClientsFrom(getClientsPath())
}

func saveClients(clients []Client) error {
	return saveClientsTo(getClientsPath(), clients)
}

func loadClientsFrom(path string) ([]Client, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Client{}, nil
		}
		return nil, err
	}

	var f clientsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return f.Clients, nil
}

func saveClientsTo(path string, clients []Client) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(clientsFile{Clients: clients}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
