---
page_title: "atlanticnet_dns_zone Resource - terraform-provider-atlanticnet"
subcategory: ""
description: |-
  Manages an Atlantic.Net DNS zone for a domain.
---

# atlanticnet_dns_zone (Resource)

Manages an Atlantic.Net DNS zone for a domain. DNS zones contain DNS records that route traffic to your servers.

## Example Usage

```terraform
resource "atlanticnet_dns_zone" "example" {
  name = "example.com"
}

resource "atlanticnet_dns_record" "www" {
  zone_id = atlanticnet_dns_zone.example.id
  type    = "A"
  host    = "www"
  data    = "192.0.2.1"
  ttl     = "3600"
}
```

## Argument Reference

- `name` - (Required) Domain name (e.g. `example.com`)

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` - The unique identifier for the DNS zone

## Import

Import a DNS zone by its ID or domain name:

```shell
terraform import atlanticnet_dns_zone.example example.com
```
