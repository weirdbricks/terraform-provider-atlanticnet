#!/bin/bash
# Test different parameter names for image in run-instance API
set -euo pipefail

if [ -z "${ATLANTICNET_ACCESS_KEY:-}" ]; then
  echo "Error: ATLANTICNET_ACCESS_KEY not set"
  exit 1
fi

API_URL="https://cloudapi.atlantic.net/"
ACTION="run-instance"
TIMESTAMP=$(date +%s)
RNDGUID=$(uuidgen)

# Create signature (HMAC-SHA256)
MSG="${TIMESTAMP}${RNDGUID}"
SIGNATURE=$(echo -n "$MSG" | openssl dgst -sha256 -hmac "$ATLANTICNET_PRIVATE_KEY" -binary | base64)

# Test different parameter names
PARAMS_TO_TEST=(
  "vm_image=Ubuntu-22.04_64bit"
  "image=Ubuntu-22.04_64bit"
  "imageid=Ubuntu-22.04_64bit"
  "osid=Ubuntu-22.04_64bit"
  "os_id=Ubuntu-22.04_64bit"
  "os_type_id=Ubuntu-22.04_64bit"
)

for PARAM in "${PARAMS_TO_TEST[@]}"; do
  echo ""
  echo "=== Testing with parameter: $PARAM ==="

  TIMESTAMP=$(date +%s)
  RNDGUID=$(uuidgen)
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
  QUERY="${QUERY}&servername=test-param-$(date +%s)"
  QUERY="${QUERY}&planname=G2.1GB"
  QUERY="${QUERY}&vm_location=USEAST2"
  QUERY="${QUERY}&server_qty=1"
  QUERY="${QUERY}&enablebackup=N"
  QUERY="${QUERY}&term=on-demand"
  QUERY="${QUERY}&${PARAM}"

  echo "Parameter: $PARAM"
  echo "Request: ${API_URL}?${QUERY}"
  echo ""

  RESPONSE=$(curl -s "${API_URL}?${QUERY}")

  echo "Full response:"
  echo "$RESPONSE" | jq .
  echo ""

  # Check if there's an error
  if echo "$RESPONSE" | jq -e '.error' > /dev/null 2>&1; then
    ERROR=$(echo "$RESPONSE" | jq -r '.error.message')
    echo "ERROR: $ERROR"
  else
    # Extract instance ID - try different paths
    INSTANCE_ID=$(echo "$RESPONSE" | jq -r '.["run-instanceresponse"]["instancesSet"]["item"]["InstanceId"] // .["run-instanceresponse"]["instancesSet"]["item"]["vm_id"] // empty' 2>/dev/null || echo "")

    if [ -n "$INSTANCE_ID" ] && [ "$INSTANCE_ID" != "null" ]; then
      echo "SUCCESS - Instance ID: $INSTANCE_ID"
      echo "Instance created! Check the UI to see which OS was created."
      exit 0
    else
      echo "Could not extract instance ID from response"
    fi
  fi
done

echo ""
echo "No successful parameter found"
