#!/bin/bash
# Check the status of the most recently created instance
set -euo pipefail

# Decrypt secrets to get credentials
SECRETS_FILE="/home/lampros/git_work/dirless-infra/secrets.env"
if [ -f "$SECRETS_FILE" ]; then
  source <(grep ATLANTICNET "$SECRETS_FILE" | grep -v '^#' | sed 's/encrypted://g')
fi

if [ -z "${ATLANTICNET_ACCESS_KEY:-}" ]; then
  echo "Error: ATLANTICNET_ACCESS_KEY not set"
  exit 1
fi

echo "Checking for instances..."

# Use terraform to show the instance details
cd /tmp/atlanticnet-instance-test
export TF_CLI_CONFIG_FILE=.terraformrc

if [ -f terraform.tfstate ]; then
  echo "Instance details from state:"
  terraform state show -json 2>/dev/null | jq '.values.root_module.resources[0].instances[0].attributes' 2>/dev/null | head -40
fi
