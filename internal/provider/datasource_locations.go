package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/user/terraform-provider-atlanticnet/internal/client"
)

var _ datasource.DataSource = &LocationsDataSource{}

func NewLocationsDataSource() datasource.DataSource { return &LocationsDataSource{} }

type LocationsDataSource struct{ client *client.Client }

type locationsModel struct {
	ID        types.String    `tfsdk:"id"`
	Locations []locationModel `tfsdk:"locations"`
}

type locationModel struct {
	Code        types.String `tfsdk:"code"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsActive    types.String `tfsdk:"is_active"`
}

func (d *LocationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locations"
}

func (d *LocationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all available Atlantic.Net datacenter locations. Use `code` as the `vm_location` argument in `atlanticnet_server`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"locations": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code":        schema.StringAttribute{Computed: true, MarkdownDescription: "Location code to use in resources (e.g. `USEAST2`)."},
						"name":        schema.StringAttribute{Computed: true, MarkdownDescription: "Short name (e.g. `USA-East-2`)."},
						"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Full description including city."},
						"is_active":   schema.StringAttribute{Computed: true, MarkdownDescription: "`Y` if accepting new servers."},
					},
				},
			},
		},
	}
}

func (d *LocationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *LocationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	locs, err := d.client.ListLocations()
	if err != nil {
		resp.Diagnostics.AddError("Failed to read locations", err.Error())
		return
	}

	data := locationsModel{ID: types.StringValue("locations")}
	for _, l := range locs {
		data.Locations = append(data.Locations, locationModel{
			Code:        types.StringValue(l.Code),
			Name:        types.StringValue(l.Name),
			Description: types.StringValue(l.Description),
			IsActive:    types.StringValue(l.IsActive),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
