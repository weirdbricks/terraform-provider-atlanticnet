---
page_title: "atlanticnet_block_volume Resource - terraform-provider-atlanticnet"
subcategory: ""
description: |-
  Manages an Atlantic.Net block storage volume.
---

# atlanticnet_block_volume (Resource)

Manages an Atlantic.Net block storage volume. Volumes can be attached to servers for additional storage.

## Example Usage

```terraform
resource "atlanticnet_server" "web" {
  name        = "web-01"
  plan_name   = "G2.4GB"
  image_id    = "Ubuntu-22.04_64bit"
  vm_location = "USEAST2"
}

resource "atlanticnet_block_volume" "data" {
  name        = "web-data"
  size_gb     = 100
  location    = "USEAST2"
  instance_id = atlanticnet_server.web.id
}
```

## Argument Reference

- `name` - (Required) Volume name/label
- `size_gb` - (Required) Volume size in GB (minimum 50, increments of 50)
- `location` - (Required) Datacenter code (must match the attached server's location)
- `instance_id` - (Optional) Cloud Server ID to attach the volume to. Set to `null` to detach.

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` - The unique identifier for the block volume
- `status` - Volume status (e.g. `AVAILABLE`, `ATTACHED`)

## Import

Import a block volume by its ID:

```shell
terraform import atlanticnet_block_volume.data volume-id
```
