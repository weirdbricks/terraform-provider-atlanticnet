#!/bin/bash
# Query an existing instance to see what the API returns
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <instance-id>"
  echo "Example: $0 i-12345678"
  exit 1
fi

INSTANCE_ID="$1"
PROVIDER_DIR="$(cd "$(dirname "$0")/.." && pwd)"

cd "$PROVIDER_DIR"
make install

mkdir -p /tmp/query-instance
cd /tmp/query-instance

cat > main.tf << 'EOF'
terraform {
  required_providers {
    atlanticnet = {
      source = "weirdbricks/atlanticnet"
    }
  }
}

provider "atlanticnet" {}

# Read the instance from state by importing it
import {
  to = atlanticnet_server.query
  id = var.instance_id
}

variable "instance_id" {
  type = string
}

output "instance_data" {
  value = atlanticnet_server.query
}
EOF

cat > .terraformrc << 'EOF'
provider_installation {
  dev_overrides {
    "weirdbricks/atlanticnet" = "/home/lampros/.terraform.d/plugins/registry.terraform.io/weirdbricks/atlanticnet/0.1.0/linux_amd64"
  }
  direct {}
}
EOF

export TF_CLI_CONFIG_FILE=.terraformrc

echo "Querying instance $INSTANCE_ID..."
terraform apply -auto-approve -var "instance_id=$INSTANCE_ID" -no-color 2>&1

echo ""
echo "=== Full instance data ==="
terraform output -json instance_data | jq .
