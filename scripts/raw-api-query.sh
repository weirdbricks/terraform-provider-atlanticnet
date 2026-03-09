#!/bin/bash
# Query Atlantic.Net API directly using curl
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <instance-id>"
  echo "Example: $0 2373463"
  exit 1
fi

INSTANCE_ID="$1"

# Get credentials from environment or secrets file
if [ -z "${ATLANTICNET_ACCESS_KEY:-}" ]; then
  if [ -f "/home/lampros/git_work/dirless-infra/secrets.env" ]; then
    # Read the encrypted secrets and extract the keys
    # Note: This requires werk vault to be configured, so we'll use a simpler approach
    echo "Error: ATLANTICNET_ACCESS_KEY not in environment"
    echo "Please set it first: source <(grep ATLANTICNET /home/lampros/git_work/dirless-infra/secrets.env | grep -v '^#')"
    exit 1
  fi
fi

API_URL="https://cloudapi.atlantic.net/"
ACTION="describe-instance"
TIMESTAMP=$(date +%s)
RNDGUID=$(uuidgen)

# Create signature (HMAC-SHA256)
MSG="${TIMESTAMP}${RNDGUID}"
SIGNATURE=$(echo -n "$MSG" | openssl dgst -sha256 -hmac "$ATLANTICNET_PRIVATE_KEY" -binary | base64)

# Build query string
QUERY="Action=${ACTION}"
QUERY="${QUERY}&Format=json"
QUERY="${QUERY}&Version=2010-12-30"
QUERY="${QUERY}&ACSAccessKeyId=${ATLANTICNET_ACCESS_KEY}"
QUERY="${QUERY}&Timestamp=${TIMESTAMP}"
QUERY="${QUERY}&Rndguid=${RNDGUID}"
QUERY="${QUERY}&Signature=$(echo -n "$SIGNATURE" | jq -sRr @uri)"
QUERY="${QUERY}&instanceid=${INSTANCE_ID}"

echo "Querying Atlantic.Net API for instance ${INSTANCE_ID}..."
echo "URL: ${API_URL}?${QUERY}"
echo ""

RESPONSE=$(curl -s "${API_URL}?${QUERY}")

echo "=== Raw API Response ==="
echo "$RESPONSE" | jq .

echo ""
echo "=== Instance Details (from describe-instance response) ==="
echo "$RESPONSE" | jq '.["describe-instanceresponse"]["instanceSet"]["item"]' 2>/dev/null || echo "Could not parse instance data"
