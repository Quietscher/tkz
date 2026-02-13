package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	},
}

// DiscoverOIDC fetches the OpenID Connect configuration from the issuer URL
func DiscoverOIDC(issuerURL string) (*OIDCConfig, error) {
	if !strings.HasPrefix(issuerURL, "https://") {
		return nil, fmt.Errorf("issuer URL must use HTTPS: %s", issuerURL)
	}
	wellKnownURL := strings.TrimRight(issuerURL, "/") + "/.well-known/openid-configuration"

	resp, err := httpClient.Get(wellKnownURL)
	if err != nil {
		return nil, fmt.Errorf("OIDC discovery request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery returned status %d", resp.StatusCode)
	}

	var config OIDCConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse OIDC discovery response: %w", err)
	}

	if config.TokenEndpoint == "" {
		return nil, fmt.Errorf("OIDC discovery response missing token_endpoint")
	}

	return &config, nil
}

// RequestToken performs a client_credentials grant against the token endpoint
func RequestToken(tokenEndpoint, clientID, clientSecret, scopes string) (*TokenResponse, error) {
	if !strings.HasPrefix(tokenEndpoint, "https://") {
		return nil, fmt.Errorf("token endpoint must use HTTPS: %s", tokenEndpoint)
	}
	data := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}
	if scopes != "" {
		data.Set("scope", scopes)
	}

	resp, err := httpClient.PostForm(tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned status %d: %s", resp.StatusCode, string(body))
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &token, nil
}
