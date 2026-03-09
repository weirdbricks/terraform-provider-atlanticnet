#!/bin/bash
# Debug: Show raw API response for an instance
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <instance-id>"
  exit 1
fi

INSTANCE_ID="$1"
PROVIDER_DIR="$(cd "$(dirname "$0")/.." && pwd)"

cd "$PROVIDER_DIR"
make install

# Export credentials from secrets
source <(grep ATLANTICNET /home/lampros/git_work/dirless-infra/secrets.env | grep -v '^#')

# Create a simple Terraform config that just reads the instance
cat > /tmp/debug-instance/main.tf << 'EOF'
terraform {
  required_providers {
    atlanticnet = {
      source = "weirdbricks/atlanticnet"
    }
  }
}

provider "atlanticnet" {}

# Read the specific instance
data "atlanticnet_server" "debug" {
  id = var.instance_id
}

variable "instance_id" {
  type = string
}

output "instance" {
  value = data.atlanticnet_server.debug
}
EOF

cat > /tmp/debug-instance/.terraformrc << 'EOF'
provider_installation {
  dev_overrides {
    "weirdbricks/atlanticnet" = "/home/lampros/.terraform.d/plugins/registry.terraform.io/weirdbricks/atlanticnet/0.1.0/linux_amd64"
  }
  direct {}
}
EOF

mkdir -p /tmp/debug-instance
cd /tmp/debug-instance
export TF_CLI_CONFIG_FILE=.terraformrc

terraform apply -auto-approve -var "instance_id=$INSTANCE_ID" -no-color

echo ""
echo "Instance details:"
terraform output -json instance | jq .
