package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/requestflo/scalegrid-terraform/internal/client"
)

var (
	_ datasource.DataSource              = (*clusterDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*clusterDataSource)(nil)
)

// NewClusterDataSource is the constructor registered with the provider.
func NewClusterDataSource() datasource.DataSource {
	return &clusterDataSource{}
}

type clusterDataSource struct {
	client *client.Client
}

type clusterDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	DatabaseType     types.String `tfsdk:"database_type"`
	Version          types.String `tfsdk:"version"`
	DeploymentType   types.String `tfsdk:"deployment_type"`
	CloudProfileID   types.String `tfsdk:"cloud_profile_id"`
	Region           types.String `tfsdk:"region"`
	SizeID           types.String `tfsdk:"size_id"`
	DiskSizeGB       types.Int64  `tfsdk:"disk_size_gb"`
	Status           types.String `tfsdk:"status"`
	Host             types.String `tfsdk:"host"`
	Port             types.Int64  `tfsdk:"port"`
	ConnectionString types.String `tfsdk:"connection_string"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

func (d *clusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *clusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single ScaleGrid cluster by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the cluster to look up.",
			},
			"name":              schema.StringAttribute{Computed: true, Description: "Name of the cluster."},
			"database_type":     schema.StringAttribute{Computed: true, Description: "Database engine."},
			"version":           schema.StringAttribute{Computed: true, Description: "Engine version."},
			"deployment_type":   schema.StringAttribute{Computed: true, Description: "Cluster topology."},
			"cloud_profile_id":  schema.StringAttribute{Computed: true, Description: "Cloud profile ID."},
			"region":            schema.StringAttribute{Computed: true, Description: "Cloud region."},
			"size_id":           schema.StringAttribute{Computed: true, Description: "Instance size identifier."},
			"disk_size_gb":      schema.Int64Attribute{Computed: true, Description: "Disk size in GB."},
			"status":            schema.StringAttribute{Computed: true, Description: "Lifecycle status."},
			"host":              schema.StringAttribute{Computed: true, Description: "Primary hostname."},
			"port":              schema.Int64Attribute{Computed: true, Description: "Connection port."},
			"connection_string": schema.StringAttribute{Computed: true, Sensitive: true, Description: "Connection string."},
			"created_at":        schema.StringAttribute{Computed: true, Description: "Creation timestamp."},
		},
	}
}

func (d *clusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, err := clientFromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected provider data", err.Error())
		return
	}
	d.client = c
}

func (d *clusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config clusterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := d.client.GetCluster(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading cluster", err.Error())
		return
	}

	state := clusterToDataSourceModel(cluster)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func clusterToDataSourceModel(cluster *client.Cluster) clusterDataSourceModel {
	return clusterDataSourceModel{
		ID:               types.StringValue(cluster.ID),
		Name:             types.StringValue(cluster.Name),
		DatabaseType:     types.StringValue(cluster.DatabaseType),
		Version:          optionalString(cluster.Version),
		DeploymentType:   optionalString(cluster.DeploymentType),
		CloudProfileID:   optionalString(cluster.CloudProfileID),
		Region:           optionalString(cluster.Region),
		SizeID:           optionalString(cluster.SizeID),
		DiskSizeGB:       types.Int64Value(cluster.DiskSizeGB),
		Status:           optionalString(cluster.Status),
		Host:             optionalString(cluster.Host),
		Port:             types.Int64Value(cluster.Port),
		ConnectionString: optionalString(cluster.ConnectionString),
		CreatedAt:        optionalString(cluster.CreatedAt),
	}
}
