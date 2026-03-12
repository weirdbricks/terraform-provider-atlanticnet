---
page_title: "atlanticnet_server Resource - terraform-provider-atlanticnet"
subcategory: ""
description: |-
  Manages an Atlantic.Net Cloud Server instance.
---

# atlanticnet_server (Resource)

Manages an Atlantic.Net Cloud Server instance. Supports creation, resizing, and deletion of servers.

## Example Usage

```terraform
resource "atlanticnet_ssh_key" "deployer" {
  name       = "deployer"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "atlanticnet_server" "web" {
  name        = "web-01"
  plan_name   = "G2.4GB"
  image_id    = "Ubuntu-22.04_64bit"
  vm_location = "USEAST2"
  ssh_key_id  = atlanticnet_ssh_key.deployer.id
}
```

## Argument Reference

- `name` - (Required) Hostname / server description
- `plan_name` - (Required) Plan size (e.g. `G2.4GB`). Can be resized to a larger plan in-place.
- `image_id` - (Required) OS image ID (e.g. `Ubuntu-22.04_64bit`). Use the `atlanticnet_images` data source to list available images.
- `vm_location` - (Required) Datacenter code (e.g. `USEAST2`). Use the `atlanticnet_locations` data source to list available regions.
- `ssh_key_id` - (Optional) SSH key ID to embed at server creation
- `enable_backup` - (Optional) Enable automated backups (default: `false`)
- `term` - (Optional) Billing term: `on-demand`, `1-year`, or `3-year` (default: `on-demand`)

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` - The unique identifier for the server
- `ip_address` - The public IPv4 address
- `status` - Server status (e.g. `RUNNING`, `PROVISIONED`)
- `cpu` - Number of CPU cores
- `ram` - RAM in GB
- `disk` - Disk space in GB
- `rate_per_hr` - Hourly billing rate
- `created_date` - Server creation timestamp

## Import

Import a server by its ID:

```shell
terraform import atlanticnet_server.web 153979
```
