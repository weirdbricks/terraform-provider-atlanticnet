---
page_title: "atlanticnet_plans Data Source - terraform-provider-atlanticnet"
subcategory: ""
description: |-
  Retrieve available Atlantic.Net server plans with pricing and specifications.
---

# atlanticnet_plans (Data Source)

Retrieves a list of available Atlantic.Net server plans with their specifications and pricing.

## Example Usage

```terraform
data "atlanticnet_plans" "available" {}

output "plan_names" {
  value = [for plan in data.atlanticnet_plans.available.plans : plan.name]
}

# Find a plan with at least 4GB RAM
output "large_plans" {
  value = [for plan in data.atlanticnet_plans.available.plans :
    plan if plan.ram >= 4
  ]
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

- `plans` - List of available plans with the following attributes:
  - `name` - Plan name (e.g. `G2.4GB`)
  - `cpu` - Number of CPU cores
  - `ram` - RAM in GB
  - `disk` - Disk space in GB
  - `rate_per_hr` - Hourly billing rate
  - `rate_per_month` - Monthly billing rate
