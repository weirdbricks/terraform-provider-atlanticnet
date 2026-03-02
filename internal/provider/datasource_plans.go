package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/user/terraform-provider-atlanticnet/internal/client"
)

var _ datasource.DataSource = &PlansDataSource{}

func NewPlansDataSource() datasource.DataSource { return &PlansDataSource{} }

type PlansDataSource struct{ client *client.Client }

type plansModel struct {
	ID    types.String `tfsdk:"id"`
	Plans []planModel  `tfsdk:"plans"`
}

type planModel struct {
	Name      types.String `tfsdk:"name"`
	RAM       types.String `tfsdk:"ram"`
	Disk      types.String `tfsdk:"disk"`
	CPU       types.String `tfsdk:"cpu"`
	RatePerHr types.String `tfsdk:"rate_per_hr"`
}

func (d *PlansDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plans"
}

func (d *PlansDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all available Atlantic.Net server plans with their specs and pricing.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"plans": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":       schema.StringAttribute{Computed: true, MarkdownDescription: "Plan name to use as `plan_name` in `atlanticnet_server`."},
						"ram":        schema.StringAttribute{Computed: true, MarkdownDescription: "RAM in MB."},
						"disk":       schema.StringAttribute{Computed: true, MarkdownDescription: "Disk size in GB."},
						"cpu":        schema.StringAttribute{Computed: true, MarkdownDescription: "Number of vCPUs."},
						"rate_per_hr": schema.StringAttribute{Computed: true, MarkdownDescription: "Hourly rate in USD."},
					},
				},
			},
		},
	}
}

func (d *PlansDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PlansDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	plans, err := d.client.ListPlans()
	if err != nil {
		resp.Diagnostics.AddError("Failed to read plans", err.Error())
		return
	}

	data := plansModel{ID: types.StringValue("plans")}
	for _, p := range plans {
		data.Plans = append(data.Plans, planModel{
			Name:      types.StringValue(p.Name),
			RAM:       types.StringValue(p.RAM),
			Disk:      types.StringValue(p.Disk),
			CPU:       types.StringValue(p.CPU),
			RatePerHr: types.StringValue(p.RatePerHr),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
