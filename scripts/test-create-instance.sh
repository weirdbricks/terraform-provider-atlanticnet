#!/bin/bash
# Create and destroy an Atlantic.Net test instance
set -euo pipefail

PROVIDER_DIR="$(cd "$(dirname "$0")/.." && pwd)"
TEST_DIR="/tmp/atlanticnet-instance-test"

echo "=== Building provider ==="
cd "$PROVIDER_DIR"
make install

echo ""
echo "=== Setting up test environment ==="
mkdir -p "$TEST_DIR"

cat > "$TEST_DIR/main.tf" << 'EOF'
terraform {
  required_providers {
    atlanticnet = {
      source = "weirdbricks/atlanticnet"
    }
  }
}

provider "atlanticnet" {}

# Create a test server
resource "atlanticnet_server" "test" {
  name        = "terraform-test-${substr(timestamp(), 0, 10)}"
  plan_name   = "G2.1GB"
  image_id    = "ubuntu-22.04_64bit"
  vm_location = "USEAST2"
  term        = "on-demand"

  enable_backup = false
}

output "server_id" {
  description = "Server ID"
  value       = atlanticnet_server.test.id
}

output "server_ip" {
  description = "Public IP address"
  value       = atlanticnet_server.test.ip_address
}

output "server_name" {
  description = "Server name"
  value       = atlanticnet_server.test.name
}
EOF

cat > "$TEST_DIR/.terraformrc" << 'EOF'
provider_installation {
  dev_overrides {
    "weirdbricks/atlanticnet" = "/home/lampros/.terraform.d/plugins/registry.terraform.io/weirdbricks/atlanticnet/0.1.0/linux_amd64"
  }
  direct {}
}
EOF

echo ""
echo "=== Running terraform plan ==="
cd "$TEST_DIR"
export TF_CLI_CONFIG_FILE=.terraformrc
terraform plan -no-color

echo ""
echo "=== Running terraform apply ==="
terraform apply -auto-approve -no-color

echo ""
echo "Instance created! Details:"
terraform output -no-color

echo ""
echo "Waiting 10 seconds before destroying..."
sleep 10

echo ""
echo "=== Running terraform destroy ==="
terraform destroy -auto-approve -no-color

echo "=== Instance destroyed ==="
