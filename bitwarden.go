package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// CheckBWInstalled checks if the bw CLI is on the PATH
func CheckBWInstalled() bool {
	_, err := exec.LookPath("bw")
	return err == nil
}

// CheckBWStatus returns the vault status: "unlocked", "locked", or "unauthenticated"
func CheckBWStatusDetail(session string) string {
	args := []string{"status"}
	if session != "" {
		args = append(args, "--session", session)
	}
	cmd := exec.Command("bw", args...)
	output, err := cmd.Output()
	if err != nil {
		return "unauthenticated"
	}
	status, err := parseBWStatusDetail(output)
	if err != nil {
		return "unauthenticated"
	}
	return status
}

// UnlockBWVault unlocks the vault with a master password and returns the session token
func UnlockBWVault(password string) (string, error) {
	cmd := exec.Command("bw", "unlock", "--passwordfile", "/dev/stdin", "--raw")
	cmd.Stdin = strings.NewReader(password)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("unlock failed: %s", errMsg)
	}

	session := strings.TrimSpace(stdout.String())
	if session == "" {
		return "", fmt.Errorf("no session token returned")
	}
	return session, nil
}

// FetchBWItems lists items from the vault, optionally filtered by search term
func FetchBWItems(session string, search string) ([]BWItem, error) {
	args := []string{"list", "items", "--session", session}
	if search != "" {
		args = append(args, "--search", search)
	}
	cmd := exec.Command("bw", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bw list items: %s", string(output))
	}
	return parseBWItems(output)
}

// FetchBWItem gets credentials for a single Bitwarden item by ID
func FetchBWItem(session string, itemID string) (*BWCredentials, error) {
	cmd := exec.Command("bw", "get", "item", itemID, "--session", session)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bw get item: %s", string(output))
	}
	return parseBWItemCredentials(output)
}

// fetchBWRawItem gets the raw JSON output for a single Bitwarden item by ID
func fetchBWRawItem(session string, itemID string) ([]byte, error) {
	cmd := exec.Command("bw", "get", "item", itemID, "--session", session)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bw get item: %s", string(output))
	}
	return output, nil
}

// --- JSON parsing functions (tested independently) ---

func parseBWStatusDetail(data []byte) (string, error) {
	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(data, &status); err != nil {
		return "", err
	}
	return status.Status, nil
}

func parseBWItems(data []byte) ([]BWItem, error) {
	var items []BWItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func parseBWItemCredentials(data []byte) (*BWCredentials, error) {
	item, err := parseBWFullItem(data)
	if err != nil {
		return nil, err
	}
	if item.Login == nil {
		return nil, fmt.Errorf("bitwarden item has no login section")
	}

	creds := &BWCredentials{
		ClientID:     item.Login.Username,
		ClientSecret: item.Login.Password,
	}
	for _, u := range item.Login.URIs {
		creds.URIs = append(creds.URIs, u.URI)
	}
	return creds, nil
}

func parseBWFullItem(data []byte) (*BWFullItem, error) {
	var item BWFullItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// ResolveBWField resolves a field path to a value from a Bitwarden item.
// Supported paths: login.username, login.password, fields.<name>, notes, or empty string.
func ResolveBWField(item *BWFullItem, fieldPath string) (string, error) {
	if fieldPath == "" {
		return "", nil
	}

	switch {
	case fieldPath == "login.username":
		if item.Login == nil {
			return "", fmt.Errorf("bitwarden item has no login section")
		}
		return item.Login.Username, nil

	case fieldPath == "login.password":
		if item.Login == nil {
			return "", fmt.Errorf("bitwarden item has no login section")
		}
		return item.Login.Password, nil

	case fieldPath == "notes":
		return item.Notes, nil

	case strings.HasPrefix(fieldPath, "fields."):
		fieldName := strings.TrimPrefix(fieldPath, "fields.")
		for _, f := range item.Fields {
			if f.Name == fieldName {
				return f.Value, nil
			}
		}
		return "", fmt.Errorf("custom field %q not found in bitwarden item", fieldName)

	default:
		return "", fmt.Errorf("unsupported field path: %s (use login.username, login.password, fields.<name>, or notes)", fieldPath)
	}
}
