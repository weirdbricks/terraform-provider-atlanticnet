---
page_title: "atlanticnet_ssh_key Resource - terraform-provider-atlanticnet"
subcategory: ""
description: |-
  Manages an Atlantic.Net SSH key pair for server authentication.
---

# atlanticnet_ssh_key (Resource)

Manages an Atlantic.Net SSH key pair. SSH keys are used to authenticate access to cloud servers.

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

- `name` - (Required) Label / name for the SSH key
- `public_key` - (Required) Public key material in OpenSSH format

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

- `id` - The unique identifier for the SSH key
- `fingerprint` - The SSH key fingerprint (MD5 hash)

## Import

SSH keys cannot be imported (they are recreated from public key material).
