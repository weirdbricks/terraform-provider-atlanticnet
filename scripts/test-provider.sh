#!/bin/bash
# Test Atlantic.Net Terraform provider
set -euo pipefail

PROVIDER_DIR="$(cd "$(dirname "$0")/.." && pwd)"
TEST_DIR="/tmp/atlanticnet-test"

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

# Test data sources only (no resources that cost money)
data "atlanticnet_locations" "all" {}
data "atlanticnet_plans" "all" {}

output "locations" {
  description = "All available locations"
  value       = [for l in data.atlanticnet_locations.all.locations : { code = l.code, name = l.name }]
}

output "plans" {
  description = "All available plans (first 5)"
  value       = slice([for p in data.atlanticnet_plans.all.plans : { name = p.name, ram = p.ram, disk = p.disk, cpu = p.cpu, rate_per_hr = p.rate_per_hr }], 0, min(5, length(data.atlanticnet_plans.all.plans)))
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
echo "=== Test complete ==="
