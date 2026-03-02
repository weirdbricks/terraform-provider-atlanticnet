package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/user/terraform-provider-atlanticnet/internal/provider"
)

// providerFactories maps the provider name to a factory for acceptance tests.
var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"atlanticnet": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// skipIfNoAccCreds skips the test if acceptance test environment variables are not set.
func skipIfNoAccCreds(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Set TF_ACC=1 and ATLANTICNET_ACCESS_KEY / ATLANTICNET_PRIVATE_KEY to run acceptance tests")
	}
	if os.Getenv("ATLANTICNET_ACCESS_KEY") == "" || os.Getenv("ATLANTICNET_PRIVATE_KEY") == "" {
		t.Skip("ATLANTICNET_ACCESS_KEY and ATLANTICNET_PRIVATE_KEY must be set for acceptance tests")
	}
}

// ─── Provider configuration ──────────────────────────────────────────────────

func TestAccProvider_basic(t *testing.T) {
	skipIfNoAccCreds(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `provider "atlanticnet" {}`,
			},
		},
	})
}

// ─── Data Sources ─────────────────────────────────────────────────────────────

func TestAccDataSourceLocations(t *testing.T) {
	skipIfNoAccCreds(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "atlanticnet_locations" "all" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlanticnet_locations.all", "locations.#"),
				),
			},
		},
	})
}

func TestAccDataSourcePlans(t *testing.T) {
	skipIfNoAccCreds(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "atlanticnet_plans" "all" {}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.atlanticnet_plans.all", "plans.#"),
				),
			},
		},
	})
}

// ─── SSH Key ──────────────────────────────────────────────────────────────────

func TestAccSSHKey_basic(t *testing.T) {
	skipIfNoAccCreds(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "atlanticnet_ssh_key" "test" {
  name       = "tf-acc-test-key"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC0+test+key+for+acc+tests tf-acc-test"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("atlanticnet_ssh_key.test", "name", "tf-acc-test-key"),
					resource.TestCheckResourceAttrSet("atlanticnet_ssh_key.test", "id"),
					resource.TestCheckResourceAttrSet("atlanticnet_ssh_key.test", "fingerprint"),
				),
			},
		},
	})
}

// ─── DNS Zone and Record ──────────────────────────────────────────────────────

func TestAccDNSZone_basic(t *testing.T) {
	skipIfNoAccCreds(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "atlanticnet_dns_zone" "test" {
  name = "tf-acc-test.example.com"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("atlanticnet_dns_zone.test", "name", "tf-acc-test.example.com"),
					resource.TestCheckResourceAttrSet("atlanticnet_dns_zone.test", "id"),
				),
			},
		},
	})
}

func TestAccDNSRecord_lifecycle(t *testing.T) {
	skipIfNoAccCreds(t)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: `
resource "atlanticnet_dns_zone" "test" {
  name = "tf-acc-record-test.example.com"
}

resource "atlanticnet_dns_record" "www" {
  zone_id = atlanticnet_dns_zone.test.id
  type    = "A"
  host    = "www"
  data    = "1.2.3.4"
  ttl     = "3600"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("atlanticnet_dns_record.www", "type", "A"),
					resource.TestCheckResourceAttr("atlanticnet_dns_record.www", "data", "1.2.3.4"),
					resource.TestCheckResourceAttrSet("atlanticnet_dns_record.www", "id"),
				),
			},
			// Update the record in-place
			{
				Config: `
resource "atlanticnet_dns_zone" "test" {
  name = "tf-acc-record-test.example.com"
}

resource "atlanticnet_dns_record" "www" {
  zone_id = atlanticnet_dns_zone.test.id
  type    = "A"
  host    = "www"
  data    = "5.6.7.8"
  ttl     = "7200"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("atlanticnet_dns_record.www", "data", "5.6.7.8"),
					resource.TestCheckResourceAttr("atlanticnet_dns_record.www", "ttl", "7200"),
				),
			},
		},
	})
}

// ─── Cloud Server ─────────────────────────────────────────────────────────────

// NOTE: Server acceptance tests create real billable resources.
// They are guarded by an additional env var: ATLANTICNET_RUN_SERVER_TESTS=1
func TestAccServer_basic(t *testing.T) {
	skipIfNoAccCreds(t)
	if os.Getenv("ATLANTICNET_RUN_SERVER_TESTS") == "" {
		t.Skip("Set ATLANTICNET_RUN_SERVER_TESTS=1 to run server acceptance tests (creates billable resources)")
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "atlanticnet_server" "test" {
  name        = "tf-acc-test-server"
  plan_name   = "G2.2GB"
  image_id    = "ubuntu-22.04_64bit"
  vm_location = "USEAST2"
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("atlanticnet_server.test", "name", "tf-acc-test-server"),
					resource.TestCheckResourceAttr("atlanticnet_server.test", "status", "RUNNING"),
					resource.TestCheckResourceAttrSet("atlanticnet_server.test", "ip_address"),
					resource.TestCheckResourceAttrSet("atlanticnet_server.test", "id"),
				),
			},
		},
	})
}
