---
page_title: "atlanticnet_dns_record Resource - terraform-provider-atlanticnet"
subcategory: ""
description: |-
  Manages an Atlantic.Net DNS record within a zone.
---

# atlanticnet_dns_record (Resource)

Manages an Atlantic.Net DNS record within a zone. Supports A, AAAA, CNAME, MX, TXT, NS, and SRV records.

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

resource "atlanticnet_dns_record" "mail" {
  zone_id  = atlanticnet_dns_zone.example.id
  type     = "MX"
  host     = "@"
  data     = "mail.example.com"
  priority = 10
  ttl      = "3600"
}
```

## Argument Reference

- `zone_id` - (Required) ID of the parent DNS zone
- `type` - (Required) Record type: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, or `SRV`
- `host` - (Required) Subdomain name (`www`, `@` for apex, `mail`, etc.)
- `data` - (Required) Record value (e.g. IP address, hostname, text)
- `ttl` - (Required) Time-to-live in seconds
- `priority` - (Optional) Priority for MX and SRV records

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` - The unique identifier for the DNS record

## Import

Import a DNS record by zone ID and record ID:

```shell
terraform import atlanticnet_dns_record.www zone-id/record-id
```
