package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// useTLSServer swaps the package-level httpClient to trust the test server's cert
// and restores the original client when the test finishes.
func useTLSServer(t *testing.T, server *httptest.Server) {
	t.Helper()
	orig := httpClient
	httpClient = server.Client()
	httpClient.Timeout = orig.Timeout
	t.Cleanup(func() { httpClient = orig })
}

func TestDiscoverOIDC(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			t.Errorf("expected path '/.well-known/openid-configuration', got '%s'", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"issuer": "https://auth.example.com",
			"token_endpoint": "https://auth.example.com/oauth/token",
			"authorization_endpoint": "https://auth.example.com/oauth/authorize"
		}`))
	}))
	defer server.Close()
	useTLSServer(t, server)

	config, err := DiscoverOIDC(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TokenEndpoint != "https://auth.example.com/oauth/token" {
		t.Errorf("expected token_endpoint 'https://auth.example.com/oauth/token', got '%s'", config.TokenEndpoint)
	}
	if config.Issuer != "https://auth.example.com" {
		t.Errorf("expected issuer 'https://auth.example.com', got '%s'", config.Issuer)
	}
}

func TestDiscoverOIDCTrailingSlash(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/openid-configuration" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"token_endpoint": "https://auth.example.com/token", "issuer": "https://auth.example.com"}`))
	}))
	defer server.Close()
	useTLSServer(t, server)

	config, err := DiscoverOIDC(server.URL + "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.TokenEndpoint != "https://auth.example.com/token" {
		t.Errorf("expected token_endpoint, got '%s'", config.TokenEndpoint)
	}
}

func TestDiscoverOIDCNotFound(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	useTLSServer(t, server)

	_, err := DiscoverOIDC(server.URL)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestDiscoverOIDCMissingTokenEndpoint(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"issuer": "https://auth.example.com"}`))
	}))
	defer server.Close()
	useTLSServer(t, server)

	_, err := DiscoverOIDC(server.URL)
	if err == nil {
		t.Fatal("expected error for missing token_endpoint")
	}
}

func TestRequestToken(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("grant_type") != "client_credentials" {
			t.Errorf("expected grant_type 'client_credentials', got '%s'", r.FormValue("grant_type"))
		}
		if r.FormValue("client_id") != "my-id" {
			t.Errorf("expected client_id 'my-id', got '%s'", r.FormValue("client_id"))
		}
		if r.FormValue("client_secret") != "my-secret" {
			t.Errorf("expected client_secret 'my-secret', got '%s'", r.FormValue("client_secret"))
		}
		if r.FormValue("scope") != "openid profile" {
			t.Errorf("expected scope 'openid profile', got '%s'", r.FormValue("scope"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"access_token": "eyJhbGciOiJSUzI1NiJ9.test-payload.signature",
			"token_type": "Bearer",
			"expires_in": 3600,
			"scope": "openid profile"
		}`))
	}))
	defer server.Close()
	useTLSServer(t, server)

	token, err := RequestToken(server.URL, "my-id", "my-secret", "openid profile")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "eyJhbGciOiJSUzI1NiJ9.test-payload.signature" {
		t.Errorf("unexpected access_token: %s", token.AccessToken)
	}
	if token.TokenType != "Bearer" {
		t.Errorf("expected token_type 'Bearer', got '%s'", token.TokenType)
	}
	if token.ExpiresIn != 3600 {
		t.Errorf("expected expires_in 3600, got %d", token.ExpiresIn)
	}
}

func TestRequestTokenNoScopes(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("scope") != "" {
			t.Errorf("expected no scope, got '%s'", r.FormValue("scope"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token": "tok", "token_type": "Bearer", "expires_in": 300}`))
	}))
	defer server.Close()
	useTLSServer(t, server)

	token, err := RequestToken(server.URL, "id", "secret", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "tok" {
		t.Errorf("unexpected access_token: %s", token.AccessToken)
	}
}

func TestRequestTokenError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid_client", "error_description": "Client authentication failed"}`))
	}))
	defer server.Close()
	useTLSServer(t, server)

	_, err := RequestToken(server.URL, "bad-id", "bad-secret", "")
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestRequestTokenServerError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	useTLSServer(t, server)

	_, err := RequestToken(server.URL, "id", "secret", "")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestDiscoverOIDCRejectsHTTP(t *testing.T) {
	_, err := DiscoverOIDC("http://auth.example.com")
	if err == nil {
		t.Fatal("expected error for HTTP URL")
	}
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("expected HTTPS error, got: %v", err)
	}
}

func TestRequestTokenRejectsHTTP(t *testing.T) {
	_, err := RequestToken("http://auth.example.com/token", "id", "secret", "openid")
	if err == nil {
		t.Fatal("expected error for HTTP URL")
	}
	if !strings.Contains(err.Error(), "HTTPS") {
		t.Errorf("expected HTTPS error, got: %v", err)
	}
}
