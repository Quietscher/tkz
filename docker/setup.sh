#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERTS_DIR="$SCRIPT_DIR/certs"

echo "=== tkz Keycloak Test Setup ==="
echo

# --- Generate TLS certificates ---
if [ -f "$CERTS_DIR/server.pem" ] && [ -f "$CERTS_DIR/ca.pem" ]; then
    echo "Certificates already exist in $CERTS_DIR — skipping generation."
else
    echo "Generating TLS certificates..."
    mkdir -p "$CERTS_DIR"

    # Generate CA key and cert
    openssl genrsa -out "$CERTS_DIR/ca-key.pem" 4096 2>/dev/null
    openssl req -new -x509 -key "$CERTS_DIR/ca-key.pem" -sha256 \
        -subj "/CN=tkz-test-ca" \
        -days 365 \
        -out "$CERTS_DIR/ca.pem" 2>/dev/null

    # Generate server key and CSR
    openssl genrsa -out "$CERTS_DIR/server-key.pem" 4096 2>/dev/null
    openssl req -new -key "$CERTS_DIR/server-key.pem" \
        -subj "/CN=localhost" \
        -out "$CERTS_DIR/server.csr" 2>/dev/null

    # Create extensions file for SAN
    cat > "$CERTS_DIR/ext.cnf" <<EOF
[v3_req]
subjectAltName = DNS:localhost, IP:127.0.0.1
EOF

    # Sign server cert with CA
    openssl x509 -req -in "$CERTS_DIR/server.csr" \
        -CA "$CERTS_DIR/ca.pem" -CAkey "$CERTS_DIR/ca-key.pem" -CAcreateserial \
        -extfile "$CERTS_DIR/ext.cnf" -extensions v3_req \
        -days 365 \
        -sha256 \
        -out "$CERTS_DIR/server.pem" 2>/dev/null

    # Clean up temp files
    rm -f "$CERTS_DIR/server.csr" "$CERTS_DIR/ext.cnf" "$CERTS_DIR/ca.srl"

    echo "Certificates generated in $CERTS_DIR"
fi

echo

# --- Trust CA on macOS (login keychain — no sudo, easy to remove) ---
if [[ "$(uname)" == "Darwin" ]]; then
    echo "To make your system trust the test CA, it will be added to your login keychain."
    echo "This does NOT require sudo and can be removed with ./teardown.sh"
    read -rp "Trust the CA certificate? [y/N] " answer
    if [[ "$answer" =~ ^[Yy]$ ]]; then
        security add-trusted-cert -r trustRoot \
            -k ~/Library/Keychains/login.keychain-db "$CERTS_DIR/ca.pem"
        echo "CA trusted (login keychain). Remove with: ./docker/teardown.sh"
    else
        echo "Skipped. You may need to trust the CA manually or use --insecure flags."
    fi
else
    echo "Not on macOS — please trust $CERTS_DIR/ca.pem in your system's certificate store manually."
fi

echo

# --- Start Keycloak ---
echo "Starting Keycloak..."
docker compose -f "$SCRIPT_DIR/docker-compose.yml" up -d
echo

echo "=== Setup Complete ==="
echo
echo "Keycloak is starting up (may take ~30s)."
echo "Check status: docker compose -f $SCRIPT_DIR/docker-compose.yml logs -f"
echo
echo "Test configuration for tkz:"
echo "  Issuer:        https://localhost:8443/realms/tkz-test"
echo "  Client ID:     tkz-test-client"
echo "  Client Secret: tkz-test-secret"
echo "  Scopes:        openid profile email"
echo
echo "Verify with:"
echo "  curl https://localhost:8443/realms/tkz-test/.well-known/openid-configuration"
