package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/user/terraform-provider-atlanticnet/internal/client"
	"strconv"
)

var _ resource.Resource = &BlockVolumeResource{}

func NewBlockVolumeResource() resource.Resource { return &BlockVolumeResource{} }

type BlockVolumeResource struct{ client *client.Client }

type blockVolumeModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	SizeGB     types.Int64  `tfsdk:"size_gb"`
	Location   types.String `tfsdk:"location"`
	InstanceID types.String `tfsdk:"instance_id"`
	// Computed
	Status types.String `tfsdk:"status"`
}

func (r *BlockVolumeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_volume"
}

func (r *BlockVolumeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages an Atlantic.Net Scalable Block Storage (SBS) volume.

Volumes are automatically encrypted at rest using LUKS and connected over an isolated storage network.
Minimum size is 50 GB; scaling increments are 50 GB.

Setting ` + "`instance_id`" + ` attaches the volume to a Cloud Server. Removing it detaches the volume.
A volume must be detached before it can be deleted.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique volume ID assigned by Atlantic.Net.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-friendly name for the volume.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"size_gb": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Volume size in GB. Minimum 50, must be a multiple of 50. Changing this recreates the volume.",
				Validators: []validator.Int64{
					int64validator.AtLeast(50),
				},
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"location": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Datacenter location code (must match the location of any attached server). Changing this recreates the volume.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"instance_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "ID of the Cloud Server to attach this volume to. Set to null to detach.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current volume status (e.g. `available`, `in-use`).",
			},
		},
	}
}

func (r *BlockVolumeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BlockVolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan blockVolumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating block volume", map[string]interface{}{
		"name":     plan.Name.ValueString(),
		"size_gb":  plan.SizeGB.ValueInt64(),
		"location": plan.Location.ValueString(),
	})

	vol, err := r.client.CreateBlockVolume(client.CreateBlockVolumeInput{
		Name:     plan.Name.ValueString(),
		Size:     strconv.FormatInt(plan.SizeGB.ValueInt64(), 10),
		Location: plan.Location.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create block volume", err.Error())
		return
	}

	plan.ID = types.StringValue(vol.ID)
	plan.Status = types.StringValue(vol.Status)

	// Attach if instance_id is set
	if !plan.InstanceID.IsNull() && plan.InstanceID.ValueString() != "" {
		tflog.Info(ctx, "Attaching block volume", map[string]interface{}{
			"volume_id":   vol.ID,
			"instance_id": plan.InstanceID.ValueString(),
		})
		if err := r.client.AttachBlockVolume(vol.ID, plan.InstanceID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to attach block volume", err.Error())
			return
		}
		plan.Status = types.StringValue("in-use")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BlockVolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state blockVolumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vol, err := r.client.GetBlockVolume(state.ID.ValueString())
	if err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Status = types.StringValue(vol.Status)
	if vol.InstanceID != "" {
		state.InstanceID = types.StringValue(vol.InstanceID)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BlockVolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state blockVolumeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planInstance := plan.InstanceID.ValueString()
	stateInstance := state.InstanceID.ValueString()

	switch {
	case planInstance == "" && stateInstance != "":
		// Detach
		tflog.Info(ctx, "Detaching block volume", map[string]interface{}{"volume_id": state.ID.ValueString()})
		if err := r.client.DetachBlockVolume(state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to detach block volume", err.Error())
			return
		}
		plan.Status = types.StringValue("available")

	case planInstance != "" && stateInstance == "":
		// Attach
		tflog.Info(ctx, "Attaching block volume", map[string]interface{}{
			"volume_id":   state.ID.ValueString(),
			"instance_id": planInstance,
		})
		if err := r.client.AttachBlockVolume(state.ID.ValueString(), planInstance); err != nil {
			resp.Diagnostics.AddError("Failed to attach block volume", err.Error())
			return
		}
		plan.Status = types.StringValue("in-use")

	case planInstance != "" && stateInstance != "" && planInstance != stateInstance:
		// Move to a different server: detach then attach
		if err := r.client.DetachBlockVolume(state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to detach block volume before move", err.Error())
			return
		}
		if err := r.client.AttachBlockVolume(state.ID.ValueString(), planInstance); err != nil {
			resp.Diagnostics.AddError("Failed to attach block volume to new server", err.Error())
			return
		}
		plan.Status = types.StringValue("in-use")
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BlockVolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state blockVolumeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Detach first if attached
	if !state.InstanceID.IsNull() && state.InstanceID.ValueString() != "" {
		tflog.Info(ctx, "Detaching block volume before delete", map[string]interface{}{"volume_id": state.ID.ValueString()})
		if err := r.client.DetachBlockVolume(state.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to detach block volume before deletion", err.Error())
			return
		}
	}

	tflog.Info(ctx, "Deleting block volume", map[string]interface{}{"id": state.ID.ValueString()})
	if err := r.client.DeleteBlockVolume(state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete block volume", err.Error())
	}
}
