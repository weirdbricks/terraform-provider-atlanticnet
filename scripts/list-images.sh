#!/bin/bash
# List available Atlantic.Net images
set -euo pipefail

if [ -z "${ATLANTICNET_ACCESS_KEY:-}" ] || [ -z "${ATLANTICNET_PRIVATE_KEY:-}" ]; then
  echo "Error: ATLANTICNET_ACCESS_KEY and ATLANTICNET_PRIVATE_KEY must be set"
  exit 1
fi

# Atlantic.Net API endpoint
API_URL="https://api.atlantic.net"

# Get list of operating systems
curl -s -X GET \
  -H "Authorization: Bearer $ATLANTICNET_ACCESS_KEY:$ATLANTICNET_PRIVATE_KEY" \
  "$API_URL/operating-systems" | jq '.operating_systems[] | {id, name, architecture}' | head -50
