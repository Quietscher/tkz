#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERTS_DIR="$SCRIPT_DIR/certs"

echo "=== tkz Keycloak Teardown ==="
echo

# --- Stop Keycloak ---
echo "Stopping Keycloak..."
docker compose -f "$SCRIPT_DIR/docker-compose.yml" down
echo

# --- Remove CA trust from macOS login keychain ---
if [[ "$(uname)" == "Darwin" ]]; then
    if security find-certificate -c "tkz-test-ca" ~/Library/Keychains/login.keychain-db >/dev/null 2>&1; then
        echo "Removing tkz-test-ca from login keychain..."
        security delete-certificate -c "tkz-test-ca" ~/Library/Keychains/login.keychain-db
        echo "CA certificate removed."
    else
        echo "No tkz-test-ca certificate found in login keychain — nothing to remove."
    fi
else
    echo "Not on macOS — please remove $CERTS_DIR/ca.pem from your system's certificate store manually."
fi

echo

# --- Optionally delete generated certs ---
if [ -d "$CERTS_DIR" ]; then
    read -rp "Delete generated certificates in $CERTS_DIR? [y/N] " answer
    if [[ "$answer" =~ ^[Yy]$ ]]; then
        rm -rf "$CERTS_DIR"
        echo "Certificates deleted."
    else
        echo "Certificates kept."
    fi
fi

echo
echo "=== Teardown Complete ==="
