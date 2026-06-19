package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/requestflo/scalegrid-terraform/internal/client"
)

var (
	_ datasource.DataSource              = (*cloudProfileDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*cloudProfileDataSource)(nil)
)

// NewCloudProfileDataSource is the constructor registered with the provider.
func NewCloudProfileDataSource() datasource.DataSource {
	return &cloudProfileDataSource{}
}

type cloudProfileDataSource struct {
	client *client.Client
}

type cloudProfileDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	CloudProvider types.String `tfsdk:"cloud_provider"`
	Region        types.String `tfsdk:"region"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func (d *cloudProfileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_profile"
}

func (d *cloudProfileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single ScaleGrid cloud profile by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the cloud profile. Either `id` or `name` must be set.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the cloud profile to look up. Either `id` or `name` must be set.",
			},
			"cloud_provider": schema.StringAttribute{Computed: true, Description: "Cloud provider."},
			"region":         schema.StringAttribute{Computed: true, Description: "Default region."},
			"created_at":     schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
		},
	}
}

func (d *cloudProfileDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, err := clientFromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected provider data", err.Error())
		return
	}
	d.client = c
}

func (d *cloudProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config cloudProfileDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := stringValue(config.ID)
	name := stringValue(config.Name)
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing lookup key", "One of `id` or `name` must be set.")
		return
	}

	var profile *client.CloudProfile
	var err error
	if id != "" {
		profile, err = d.client.GetCloudProfile(ctx, id)
		if err != nil {
			resp.Diagnostics.AddError("Error reading cloud profile", err.Error())
			return
		}
	} else {
		profile, err = d.findByName(ctx, name)
		if err != nil {
			resp.Diagnostics.AddError("Error looking up cloud profile by name", err.Error())
			return
		}
	}

	state := cloudProfileDataSourceModel{
		ID:            types.StringValue(profile.ID),
		Name:          types.StringValue(profile.Name),
		CloudProvider: types.StringValue(profile.CloudProvider),
		Region:        optionalString(profile.Region),
		CreatedAt:     optionalString(profile.CreatedAt),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *cloudProfileDataSource) findByName(ctx context.Context, name string) (*client.CloudProfile, error) {
	profiles, err := d.client.ListCloudProfiles(ctx)
	if err != nil {
		return nil, err
	}
	for i := range profiles {
		if profiles[i].Name == name {
			return &profiles[i], nil
		}
	}
	return nil, fmt.Errorf("no cloud profile found with name %q", name)
}
