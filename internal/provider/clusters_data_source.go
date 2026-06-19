package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/requestflo/scalegrid-terraform/internal/client"
)

var (
	_ datasource.DataSource              = (*clustersDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*clustersDataSource)(nil)
)

// NewClustersDataSource is the constructor registered with the provider.
func NewClustersDataSource() datasource.DataSource {
	return &clustersDataSource{}
}

type clustersDataSource struct {
	client *client.Client
}

type clustersDataSourceModel struct {
	DatabaseType types.String             `tfsdk:"database_type"`
	Clusters     []clusterDataSourceModel `tfsdk:"clusters"`
}

func (d *clustersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clusters"
}

func (d *clustersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists ScaleGrid clusters on the account, optionally filtered by database type.",
		Attributes: map[string]schema.Attribute{
			"database_type": schema.StringAttribute{
				Optional:    true,
				Description: "If set, only clusters with this database engine are returned.",
			},
			"clusters": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of clusters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                schema.StringAttribute{Computed: true, Description: "Cluster ID."},
						"name":              schema.StringAttribute{Computed: true, Description: "Cluster name."},
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
				},
			},
		},
	}
}

func (d *clustersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, err := clientFromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected provider data", err.Error())
		return
	}
	d.client = c
}

func (d *clustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config clustersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusters, err := d.client.ListClusters(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing clusters", err.Error())
		return
	}

	filter := stringValue(config.DatabaseType)
	config.Clusters = nil
	for i := range clusters {
		if filter != "" && clusters[i].DatabaseType != filter {
			continue
		}
		config.Clusters = append(config.Clusters, clusterToDataSourceModel(&clusters[i]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
