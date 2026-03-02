// Package provider implements the Terraform provider for Atlantic.Net Cloud.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/user/terraform-provider-atlanticnet/internal/client"
)

var _ provider.Provider = &AtlanticNetProvider{}

// AtlanticNetProvider is the root Terraform provider implementation.
type AtlanticNetProvider struct {
	version string
}

// providerModel maps provider HCL schema to Go.
type providerModel struct {
	AccessKey  types.String `tfsdk:"access_key"`
	PrivateKey types.String `tfsdk:"private_key"`
}

// New returns a factory function for the provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AtlanticNetProvider{version: version}
	}
}

func (p *AtlanticNetProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "atlanticnet"
	resp.Version = p.version
}

func (p *AtlanticNetProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
The **atlanticnet** provider manages resources on the [Atlantic.Net Cloud](https://www.atlantic.net) platform.

## Authentication

Credentials can be provided via the provider block or environment variables:

` + "```" + `bash
export ATLANTICNET_ACCESS_KEY="your_access_key"
export ATLANTICNET_PRIVATE_KEY="your_private_key"
` + "```",
		Attributes: map[string]schema.Attribute{
			"access_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Atlantic.Net API Access Key. Also read from `ATLANTICNET_ACCESS_KEY`.",
			},
			"private_key": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Atlantic.Net API Private Key. Also read from `ATLANTICNET_PRIVATE_KEY`.",
			},
		},
	}
}

func (p *AtlanticNetProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessKey := envOr(config.AccessKey.ValueString(), "ATLANTICNET_ACCESS_KEY")
	privateKey := envOr(config.PrivateKey.ValueString(), "ATLANTICNET_PRIVATE_KEY")

	if accessKey == "" {
		resp.Diagnostics.AddError(
			"Missing Atlantic.Net Access Key",
			"Set access_key in the provider block or the ATLANTICNET_ACCESS_KEY environment variable.",
		)
	}
	if privateKey == "" {
		resp.Diagnostics.AddError(
			"Missing Atlantic.Net Private Key",
			"Set private_key in the provider block or the ATLANTICNET_PRIVATE_KEY environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c := client.New(accessKey, privateKey)
	resp.ResourceData = c
	resp.DataSourceData = c
}

func (p *AtlanticNetProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewServerResource,
		NewSSHKeyResource,
		NewDNSZoneResource,
		NewDNSRecordResource,
		NewBlockVolumeResource,
	}
}

func (p *AtlanticNetProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewLocationsDataSource,
		NewPlansDataSource,
	}
}

// envOr returns val if non-empty, otherwise the value of the named env var.
func envOr(val, envKey string) string {
	if val != "" {
		return val
	}
	return os.Getenv(envKey)
}
