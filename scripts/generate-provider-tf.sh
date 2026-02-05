#!/bin/bash
# Script to generate a provider.tf from .env file
# By default writes to examples/provider/provider.tf. Override with OUTPUT_FILE.

# DONT CALL THIS SCRIPT DIRECTLY, USE task build-demo INSTEAD

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="$REPO_ROOT/.env"
OUTPUT_FILE="${OUTPUT_FILE:-$REPO_ROOT/examples/provider/provider.tf}"

if [ ! -f "$ENV_FILE" ]; then
    echo "Error: .env file not found at $ENV_FILE" >&2
    echo "Please create it from .env-example" >&2
    exit 1
fi

# Source the .env file to get variables
# Use set -a to automatically export all variables, then source the file
set -a
source "$ENV_FILE"
set +a

# Extract variables with defaults
API_BASE_URL="${PRIVX_API_BASE_URL:-}"
BEARER_TOKEN="${PRIVX_API_BEARER_TOKEN:-}"
OAUTH_CLIENT_ID="${PRIVX_API_OAUTH_CLIENT_ID:-}"
OAUTH_SECRET="${PRIVX_API_OAUTH_CLIENT_SECRET:-}"
CLIENT_ID="${PRIVX_API_CLIENT_ID:-}"
CLIENT_SECRET="${PRIVX_API_CLIENT_SECRET:-}"
TRUST_ANCHOR="${TRUST_ANCHOR:-}"

# Create output directory if it doesn't exist
mkdir -p "$(dirname "$OUTPUT_FILE")"

# Generate the provider.tf file
cat > "$OUTPUT_FILE" <<EOF
provider "privx" {
  api_base_url = "${API_BASE_URL}"
  /* Oauth auth can be replaced by token */
  //api_bearer_token        = "${BEARER_TOKEN}"
  api_oauth_client_id     = "${OAUTH_CLIENT_ID}"
  api_oauth_client_secret = "${OAUTH_SECRET}"
  api_client_id           = "${CLIENT_ID}"
  api_client_secret       = "${CLIENT_SECRET}"
  debug                   = false
EOF

if [ -n "$TRUST_ANCHOR" ]; then
  cat >> "$OUTPUT_FILE" <<EOF

  # Optional TLS trust anchor certificate (PEM) for PrivX API server
  trust_anchor = <<CERT_EOF
${TRUST_ANCHOR}
CERT_EOF
EOF
fi

cat >> "$OUTPUT_FILE" <<EOF
}
EOF

echo "Generated $OUTPUT_FILE from $ENV_FILE"

