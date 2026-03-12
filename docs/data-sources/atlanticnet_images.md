---
page_title: "atlanticnet_images Data Source - terraform-provider-atlanticnet"
subcategory: ""
description: |-
  Retrieve available Atlantic.Net OS images for provisioning servers.
---

# atlanticnet_images (Data Source)

Retrieves a list of available Atlantic.Net OS images that can be used when provisioning servers.

## Example Usage

```terraform
data "atlanticnet_images" "available" {}

output "linux_images" {
  value = [for img in data.atlanticnet_images.available.images :
    img if img.os_family == "linux"
  ]
}

# Find Ubuntu 22.04 image
output "ubuntu_image" {
  value = [for img in data.atlanticnet_images.available.images :
    img.id if img.os_name == "Ubuntu" && img.version == "22.04"
  ]
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `images` - List of available images with the following attributes:
  - `id` - Image ID (e.g. `Ubuntu-22.04_64bit`)
  - `displayname` - Human-readable image name
  - `ostype` - OS type
  - `os_family` - OS family (`linux`, `windows`, etc.)
  - `os_name` - Operating system name
  - `version` - OS version
  - `architecture` - Architecture (`64bit`, `32bit`)
