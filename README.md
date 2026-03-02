# terraform-provider-atlanticnet

A production-quality Terraform provider for [Atlantic.Net Cloud](https://www.atlantic.net).

[![CI](https://github.com/weirdbricks/terraform-provider-atlanticnet/actions/workflows/ci.yml/badge.svg)](https://github.com/weirdbricks/terraform-provider-atlanticnet/actions/workflows/ci.yml)

## Supported Resources

| Resource | Create | Read | Update | Delete | Import |
|---|:---:|:---:|:---:|:---:|:---:|
| `atlanticnet_server` | ✅ | ✅ | ✅ (resize) | ✅ | ✅ |
| `atlanticnet_ssh_key` | ✅ | ✅ | — | ✅ | — |
| `atlanticnet_dns_zone` | ✅ | ✅ | — | ✅ | — |
| `atlanticnet_dns_record` | ✅ | ✅ | ✅ | ✅ | — |
| `atlanticnet_block_volume` | ✅ | ✅ | ✅ (attach/detach) | ✅ | — |

## Supported Data Sources

| Data Source | Description |
|---|---|
| `atlanticnet_locations` | Lists all available datacenters |
| `atlanticnet_plans` | Lists all server plans with pricing |

## ⚠️ Firewall Note

Atlantic.Net's firewall rules are managed through the **web control panel only** — they are not exposed via the public API and therefore cannot be managed with this provider. Please configure firewall rules through [cloud.atlantic.net](https://cloud.atlantic.net).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build from source)
- An [Atlantic.Net](https://cloud.atlantic.net) account with API access enabled

## Getting Your API Keys

1. Log in to [cloud.atlantic.net](https://cloud.atlantic.net)
2. Click **API Info** in the sidebar
3. Enable API access if prompted
4. Copy your **API Access Key** and **API Private Key**

## Building & Installing Locally

```bash
git clone https://github.com/weirdbricks/terraform-provider-atlanticnet.git
cd terraform-provider-atlanticnet
go mod tidy
make install   # builds + installs to ~/.terraform.d/plugins
```

Then create a `dev.tfrc` override:

```hcl
provider_installation {
  dev_overrides {
    "weirdbricks/atlanticnet" = "/path/to/terraform-provider-atlanticnet"
  }
  direct {}
}
```

```bash
export TF_CLI_CONFIG_FILE=/path/to/dev.tfrc
```

## Authentication

Set credentials via environment variables (recommended for CI):

```bash
export ATLANTICNET_ACCESS_KEY="your_access_key"
export ATLANTICNET_PRIVATE_KEY="your_private_key"
```

Or in the provider block:

```hcl
provider "atlanticnet" {
  access_key  = "your_access_key"
  private_key = "your_private_key"  # use a variable or env var, never hardcode
}
```

## Example Usage

```hcl
terraform {
  required_providers {
    atlanticnet = {
      source  = "weirdbricks/atlanticnet"
      version = "~> 0.1"
    }
  }
}

provider "atlanticnet" {}

resource "atlanticnet_ssh_key" "deployer" {
  name       = "deployer"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "atlanticnet_server" "web" {
  name        = "web-01"
  plan_name   = "G2.4GB"
  image_id    = "ubuntu-22.04_64bit"
  vm_location = "USEAST2"
  ssh_key_id  = atlanticnet_ssh_key.deployer.id
}

resource "atlanticnet_block_volume" "data" {
  name        = "web-data"
  size_gb     = 100
  location    = "USEAST2"
  instance_id = atlanticnet_server.web.id
}

resource "atlanticnet_dns_zone" "main" {
  name = "example.com"
}

resource "atlanticnet_dns_record" "www" {
  zone_id = atlanticnet_dns_zone.main.id
  type    = "A"
  host    = "www"
  data    = atlanticnet_server.web.ip_address
  ttl     = "3600"
}
```

## Resource Reference

### `atlanticnet_server`

| Argument | Required | Description |
|---|---|---|
| `name` | Yes | Hostname / description |
| `plan_name` | Yes | Plan size (e.g. `G2.4GB`). Resize to larger plan is supported in-place. |
| `image_id` | Yes | OS image ID (e.g. `ubuntu-22.04_64bit`) |
| `vm_location` | Yes | Datacenter code (e.g. `USEAST2`) |
| `ssh_key_id` | No | SSH key to embed at creation |
| `enable_backup` | No | Enable backups (default: `false`) |
| `term` | No | `on-demand`, `1-year`, `3-year` (default: `on-demand`) |

**Computed:** `id`, `ip_address`, `status`, `cpu`, `ram`, `disk`, `rate_per_hr`, `created_date`

**Import:** `terraform import atlanticnet_server.web 153979`

### `atlanticnet_ssh_key`

| Argument | Required | Description |
|---|---|---|
| `name` | Yes | Label for the key |
| `public_key` | Yes | Public key material |

**Computed:** `id`, `fingerprint`

### `atlanticnet_dns_zone`

| Argument | Required | Description |
|---|---|---|
| `name` | Yes | Domain name (e.g. `example.com`) |

**Computed:** `id`

### `atlanticnet_dns_record`

| Argument | Required | Description |
|---|---|---|
| `zone_id` | Yes | ID of the parent zone |
| `type` | Yes | `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SRV` |
| `host` | Yes | Subdomain (`www`, `@` for apex) |
| `data` | Yes | Record value |
| `ttl` | Yes | TTL in seconds |
| `priority` | No | Priority (MX / SRV only) |

**Computed:** `id`

### `atlanticnet_block_volume`

| Argument | Required | Description |
|---|---|---|
| `name` | Yes | Volume name |
| `size_gb` | Yes | Size in GB (min 50, increments of 50) |
| `location` | Yes | Datacenter code (must match attached server's location) |
| `instance_id` | No | Cloud Server ID to attach to. Set to null to detach. |

**Computed:** `id`, `status`

## Running Tests

```bash
# Unit tests — no API key needed, uses a mock HTTP server
make test

# Acceptance tests — hits the real Atlantic.Net API
export ATLANTICNET_ACCESS_KEY="..."
export ATLANTICNET_PRIVATE_KEY="..."
make testacc

# Server acceptance tests — creates real billable VMs
make testacc-servers
```

## Publishing to the Terraform Registry

1. Update the `Address` in `main.go` to `registry.terraform.io/YOUR_ORG/atlanticnet`
2. Add a GPG signing key per the [Registry publishing guide](https://developer.hashicorp.com/terraform/registry/providers/publishing)
3. Push a version tag: `git tag v0.1.0 && git push origin v0.1.0`
4. GoReleaser (`.goreleaser.yml`) builds cross-platform binaries via GitHub Actions automatically

## Contributing

To add support for more Atlantic.Net resources:

1. Add client methods to `internal/client/client.go`
2. Add unit tests to `internal/client/client_test.go`
3. Create `internal/provider/resource_<name>.go`
4. Add acceptance tests to `internal/provider/provider_test.go`
5. Register the resource in `internal/provider/provider.go`
6. Add an example in `examples/resources/atlanticnet_<name>/resource.tf`
