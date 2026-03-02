terraform {
  required_providers {
    atlanticnet = {
      source  = "weirdbricks/atlanticnet"
      version = "~> 0.1"
    }
  }
  required_version = ">= 1.0"
}

# Credentials via environment variables (recommended):
#   export ATLANTICNET_ACCESS_KEY="your_access_key"
#   export ATLANTICNET_PRIVATE_KEY="your_private_key"
provider "atlanticnet" {}

# ─── Discover available infrastructure ───────────────────────────────────────

data "atlanticnet_locations" "all" {}
data "atlanticnet_plans" "all" {}

output "active_locations" {
  description = "Locations currently accepting new servers"
  value = [
    for l in data.atlanticnet_locations.all.locations :
    l.code if l.is_active == "Y"
  ]
}

# ─── SSH Key ──────────────────────────────────────────────────────────────────

resource "atlanticnet_ssh_key" "deployer" {
  name       = "deployer"
  public_key = file("~/.ssh/id_rsa.pub")
}

# ─── Cloud Server ─────────────────────────────────────────────────────────────

resource "atlanticnet_server" "web" {
  name        = "web-01"
  plan_name   = "G2.4GB"
  image_id    = "ubuntu-22.04_64bit"
  vm_location = "USEAST2"
  ssh_key_id  = atlanticnet_ssh_key.deployer.id

  enable_backup = true
  term          = "on-demand"
}

output "web_ip" {
  description = "Public IP address of web-01"
  value       = atlanticnet_server.web.ip_address
}

# ─── Block Storage ────────────────────────────────────────────────────────────

resource "atlanticnet_block_volume" "data_disk" {
  name      = "web-data"
  size_gb   = 100       # minimum 50, increments of 50
  location  = "USEAST2" # must match the server's location
  instance_id = atlanticnet_server.web.id
}

# ─── DNS ─────────────────────────────────────────────────────────────────────

resource "atlanticnet_dns_zone" "main" {
  name = "example.com"
}

resource "atlanticnet_dns_record" "apex" {
  zone_id = atlanticnet_dns_zone.main.id
  type    = "A"
  host    = "@"
  data    = atlanticnet_server.web.ip_address
  ttl     = "3600"
}

resource "atlanticnet_dns_record" "www" {
  zone_id = atlanticnet_dns_zone.main.id
  type    = "CNAME"
  host    = "www"
  data    = "example.com"
  ttl     = "3600"
}

resource "atlanticnet_dns_record" "mx" {
  zone_id  = atlanticnet_dns_zone.main.id
  type     = "MX"
  host     = "@"
  data     = "mail.example.com"
  ttl      = "3600"
  priority = "10"
}
