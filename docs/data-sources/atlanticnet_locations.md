---
page_title: "atlanticnet_locations Data Source - terraform-provider-atlanticnet"
subcategory: ""
description: |-
  Retrieve available Atlantic.Net datacenter locations with pricing information.
---

# atlanticnet_locations (Data Source)

Retrieves a list of available Atlantic.Net datacenters and their pricing information.

## Example Usage

```terraform
data "atlanticnet_locations" "available" {}

output "regions" {
  value = [for loc in data.atlanticnet_locations.available.locations : loc.code]
}

output "useast2_pricing" {
  value = [for loc in data.atlanticnet_locations.available.locations :
    loc if loc.code == "USEAST2"
  ]
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `locations` - List of available locations with the following attributes:
  - `code` - Datacenter code (e.g. `USEAST2`, `EUWEST1`)
  - `name` - Human-readable datacenter name
  - `country_code` - Country code
  - `city` - City name
  - `region` - Region/state name
