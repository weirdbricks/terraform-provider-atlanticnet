package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/weirdbricks/terraform-provider-atlanticnet/internal/client"
)

var _ datasource.DataSource = &ImagesDataSource{}

func NewImagesDataSource() datasource.DataSource { return &ImagesDataSource{} }

type ImagesDataSource struct{ client *client.Client }

type imagesModel struct {
	ID     types.String `tfsdk:"id"`
	Images []imageModel `tfsdk:"images"`
}

type imageModel struct {
	ID           types.String `tfsdk:"id"`
	DisplayName  types.String `tfsdk:"display_name"`
	OSType       types.String `tfsdk:"os_type"`
	OSFamily     types.String `tfsdk:"os_family"`
	Architecture types.String `tfsdk:"architecture"`
	Version      types.String `tfsdk:"version"`
}

func (d *ImagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *ImagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all available Atlantic.Net OS images that can be used with `atlanticnet_server` resources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"images": schema.ListNestedAttribute{
				Computed: true,
				MarkdownDescription: "List of available OS images.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true, MarkdownDescription: "Image ID to use as `image_id` in `atlanticnet_server` (e.g. `Ubuntu-22.04_64bit`)."},
						"display_name": schema.StringAttribute{Computed: true, MarkdownDescription: "Human-readable name for the image."},
						"os_type":      schema.StringAttribute{Computed: true, MarkdownDescription: "OS type (e.g. `linux`, `windows2k25`)."},
						"os_family":    schema.StringAttribute{Computed: true, MarkdownDescription: "OS family (e.g. `ubuntu`, `alma`, `debian`, `windows_server`)."},
						"architecture": schema.StringAttribute{Computed: true, MarkdownDescription: "CPU architecture (e.g. `amd64`, `x86_64`)."},
						"version":      schema.StringAttribute{Computed: true, MarkdownDescription: "Version or variant of the OS (e.g. `22.04 LTS`, `10.0`)."},
					},
				},
			},
		},
	}
}

func (d *ImagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ImagesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	images, err := d.client.ListImages()
	if err != nil {
		resp.Diagnostics.AddError("Failed to read images", err.Error())
		return
	}

	data := imagesModel{ID: types.StringValue("images")}
	for _, img := range images {
		data.Images = append(data.Images, imageModel{
			ID:           types.StringValue(img.ID),
			DisplayName:  types.StringValue(img.DisplayName),
			OSType:       types.StringValue(img.OSType),
			OSFamily:     types.StringValue(img.OSFamily),
			Architecture: types.StringValue(img.Architecture),
			Version:      types.StringValue(img.Version),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
