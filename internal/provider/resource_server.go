package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/user/terraform-provider-atlanticnet/internal/client"
)

var _ resource.Resource = &ServerResource{}
var _ resource.ResourceWithImportState = &ServerResource{}

// NewServerResource returns a new atlanticnet_server resource.
func NewServerResource() resource.Resource { return &ServerResource{} }

// ServerResource manages Atlantic.Net Cloud Server instances.
type ServerResource struct{ client *client.Client }

// serverModel is the Terraform state model for a Cloud Server.
type serverModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	PlanName     types.String `tfsdk:"plan_name"`
	ImageID      types.String `tfsdk:"image_id"`
	VMLocation   types.String `tfsdk:"vm_location"`
	SSHKeyID     types.String `tfsdk:"ssh_key_id"`
	EnableBackup types.Bool   `tfsdk:"enable_backup"`
	Term         types.String `tfsdk:"term"`
	// Computed
	IPAddress   types.String `tfsdk:"ip_address"`
	Status      types.String `tfsdk:"status"`
	CPU         types.String `tfsdk:"cpu"`
	RAM         types.String `tfsdk:"ram"`
	Disk        types.String `tfsdk:"disk"`
	RatePerHr   types.String `tfsdk:"rate_per_hr"`
	CreatedDate types.String `tfsdk:"created_date"`
}

func (r *ServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Atlantic.Net Cloud Server (VPS). Creating a server may take several minutes while the provider waits for it to reach `RUNNING` status.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique numeric ID of the Cloud Server.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Hostname / description of the Cloud Server.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"plan_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Server plan (e.g. `G2.4GB`). Use the `atlanticnet_plans` data source to list available plans. Changing this to a larger plan triggers an in-place resize.",
			},
			"image_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "OS image ID (e.g. `ubuntu-22.04_64bit`). Changing this recreates the server.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vm_location": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Datacenter location code (e.g. `USEAST2`). Use the `atlanticnet_locations` data source. Changing this recreates the server.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ssh_key_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of an SSH key to embed at creation time.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enable_backup": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to enable automated backups (default: `false`).",
			},
			"term": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("on-demand"),
				MarkdownDescription: "Billing term: `on-demand`, `1-year`, or `3-year`.",
				Validators: []validator.String{
					stringvalidator.OneOf("on-demand", "1-year", "3-year"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			// Computed
			"ip_address":   schema.StringAttribute{Computed: true, MarkdownDescription: "Primary public IP address."},
			"status":       schema.StringAttribute{Computed: true, MarkdownDescription: "Current server status (e.g. `RUNNING`)."},
			"cpu":          schema.StringAttribute{Computed: true, MarkdownDescription: "Number of virtual CPUs."},
			"ram":          schema.StringAttribute{Computed: true, MarkdownDescription: "RAM in MB."},
			"disk":         schema.StringAttribute{Computed: true, MarkdownDescription: "Disk size in GB."},
			"rate_per_hr":  schema.StringAttribute{Computed: true, MarkdownDescription: "Current hourly rate in USD."},
			"created_date": schema.StringAttribute{Computed: true, MarkdownDescription: "Unix timestamp of when the server was created."},
		},
	}
}

func (r *ServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got %T. Please report this issue.", req.ProviderData))
		return
	}
	r.client = c
}

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating Atlantic.Net Cloud Server", map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"plan":        plan.PlanName.ValueString(),
		"image":       plan.ImageID.ValueString(),
		"vm_location": plan.VMLocation.ValueString(),
	})

	inst, err := r.client.RunInstance(client.RunInstanceInput{
		ServerName:   plan.Name.ValueString(),
		ImageID:      plan.ImageID.ValueString(),
		PlanName:     plan.PlanName.ValueString(),
		VMLocation:   plan.VMLocation.ValueString(),
		SSHKeyID:     plan.SSHKeyID.ValueString(),
		EnableBackup: plan.EnableBackup.ValueBool(),
		Term:         plan.Term.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Cloud Server", err.Error())
		return
	}

	instanceToModel(inst, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inst, err := r.client.GetInstance(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Cloud Server", err.Error())
		return
	}
	if inst.Status == "REMOVED" {
		resp.State.RemoveResource(ctx)
		return
	}

	instanceToModel(inst, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serverModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The only in-place change the API supports is plan_name (resize to larger plan).
	if !plan.PlanName.Equal(state.PlanName) {
		tflog.Info(ctx, "Resizing Cloud Server", map[string]interface{}{
			"id":       state.ID.ValueString(),
			"old_plan": state.PlanName.ValueString(),
			"new_plan": plan.PlanName.ValueString(),
		})
		inst, err := r.client.ResizeInstance(state.ID.ValueString(), plan.PlanName.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to resize Cloud Server", err.Error())
			return
		}
		instanceToModel(inst, &plan)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Terminating Cloud Server", map[string]interface{}{"id": state.ID.ValueString()})
	if err := r.client.TerminateInstance(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to terminate Cloud Server", err.Error())
	}
}

func (r *ServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	inst, err := r.client.GetInstance(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to import Cloud Server", err.Error())
		return
	}
	var m serverModel
	instanceToModel(inst, &m)
	resp.Diagnostics.Append(resp.State.Set(ctx, &m)...)
}

// instanceToModel maps an API Instance to Terraform state.
func instanceToModel(inst *client.Instance, m *serverModel) {
	m.ID = types.StringValue(inst.ID)
	m.Name = types.StringValue(inst.Name)
	m.PlanName = types.StringValue(inst.PlanName)
	m.ImageID = types.StringValue(inst.Image)
	m.IPAddress = types.StringValue(inst.IPAddress)
	m.Status = types.StringValue(inst.Status)
	m.CPU = types.StringValue(inst.CPU)
	m.RAM = types.StringValue(inst.RAM)
	m.Disk = types.StringValue(inst.Disk)
	m.RatePerHr = types.StringValue(inst.RatePerHr)
	m.CreatedDate = types.StringValue(inst.CreatedDate)
}
