package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/user/terraform-provider-atlanticnet/internal/client"
)

var _ resource.Resource = &DNSRecordResource{}

func NewDNSRecordResource() resource.Resource { return &DNSRecordResource{} }

type DNSRecordResource struct{ client *client.Client }

type dnsRecordModel struct {
	ID       types.String `tfsdk:"id"`
	ZoneID   types.String `tfsdk:"zone_id"`
	Type     types.String `tfsdk:"type"`
	Host     types.String `tfsdk:"host"`
	Data     types.String `tfsdk:"data"`
	TTL      types.String `tfsdk:"ttl"`
	Priority types.String `tfsdk:"priority"`
}

func (r *DNSRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DNSRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a DNS record within an `atlanticnet_dns_zone`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique record ID assigned by Atlantic.Net.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"zone_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the parent DNS zone.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Record type: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SRV`.",
				Validators: []validator.String{
					stringvalidator.OneOf("A", "AAAA", "CNAME", "MX", "TXT", "NS", "SRV"),
				},
			},
			"host": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The host/subdomain portion (e.g. `www`, `@` for apex, `mail`).",
			},
			"data": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The record value (e.g. `1.2.3.4` for A, `mail.example.com` for MX).",
			},
			"ttl": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Time-to-live in seconds (e.g. `3600`).",
			},
			"priority": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Priority for MX and SRV records.",
			},
		},
	}
}

func (r *DNSRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *DNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating DNS record", map[string]interface{}{
		"zone_id": plan.ZoneID.ValueString(),
		"type":    plan.Type.ValueString(),
		"host":    plan.Host.ValueString(),
	})

	rec, err := r.client.CreateDNSRecord(client.CreateDNSRecordInput{
		ZoneID:   plan.ZoneID.ValueString(),
		Type:     plan.Type.ValueString(),
		Host:     plan.Host.ValueString(),
		Data:     plan.Data.ValueString(),
		TTL:      plan.TTL.ValueString(),
		Priority: plan.Priority.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create DNS record", err.Error())
		return
	}

	plan.ID = types.StringValue(rec.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rec, err := r.client.GetDNSRecord(state.ZoneID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Type = types.StringValue(rec.Type)
	state.Host = types.StringValue(rec.Host)
	state.Data = types.StringValue(rec.Data)
	state.TTL = types.StringValue(rec.TTL)
	state.Priority = types.StringValue(rec.Priority)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsRecordModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state dnsRecordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating DNS record", map[string]interface{}{"id": state.ID.ValueString()})

	rec, err := r.client.UpdateDNSRecord(client.CreateDNSRecordInput{
		ZoneID:   plan.ZoneID.ValueString(),
		Type:     plan.Type.ValueString(),
		Host:     plan.Host.ValueString(),
		Data:     plan.Data.ValueString(),
		TTL:      plan.TTL.ValueString(),
		Priority: plan.Priority.ValueString(),
	}, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to update DNS record", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Type = types.StringValue(rec.Type)
	plan.Host = types.StringValue(rec.Host)
	plan.Data = types.StringValue(rec.Data)
	plan.TTL = types.StringValue(rec.TTL)
	plan.Priority = types.StringValue(rec.Priority)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting DNS record", map[string]interface{}{"id": state.ID.ValueString()})
	if err := r.client.DeleteDNSRecord(state.ZoneID.ValueString(), state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete DNS record", err.Error())
	}
}
