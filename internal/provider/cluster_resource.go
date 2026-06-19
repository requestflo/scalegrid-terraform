package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/requestflo/scalegrid-terraform/internal/client"
)

var (
	_ resource.Resource                = (*clusterResource)(nil)
	_ resource.ResourceWithConfigure   = (*clusterResource)(nil)
	_ resource.ResourceWithImportState = (*clusterResource)(nil)
)

// clusterPollInterval controls how often cluster provisioning is polled.
const clusterPollInterval = 15 * time.Second

// NewClusterResource is the constructor registered with the provider.
func NewClusterResource() resource.Resource {
	return &clusterResource{}
}

type clusterResource struct {
	client *client.Client
}

// clusterResourceModel maps the resource schema to Go types.
type clusterResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	DatabaseType     types.String `tfsdk:"database_type"`
	Version          types.String `tfsdk:"version"`
	DeploymentType   types.String `tfsdk:"deployment_type"`
	CloudProfileID   types.String `tfsdk:"cloud_profile_id"`
	Region           types.String `tfsdk:"region"`
	SizeID           types.String `tfsdk:"size_id"`
	DiskSizeGB       types.Int64  `tfsdk:"disk_size_gb"`
	ShardCount       types.Int64  `tfsdk:"shard_count"`
	SSLEnabled       types.Bool   `tfsdk:"ssl_enabled"`
	EncryptionAtRest types.Bool   `tfsdk:"encryption_at_rest"`
	Tags             types.List   `tfsdk:"tags"`

	Status           types.String `tfsdk:"status"`
	Host             types.String `tfsdk:"host"`
	Port             types.Int64  `tfsdk:"port"`
	ConnectionString types.String `tfsdk:"connection_string"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *clusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ScaleGrid database deployment (cluster).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique identifier of the cluster.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name of the cluster.",
			},
			"database_type": schema.StringAttribute{
				Required:      true,
				Description:   "Database engine: `mongodb`, `redis`, `mysql`, or `postgresql`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.OneOf(
						client.DatabaseMongoDB,
						client.DatabaseRedis,
						client.DatabaseMySQL,
						client.DatabasePostgreSQL,
					),
				},
			},
			"version": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Engine version to deploy (for example `7.0`). Defaults to the ScaleGrid recommended version.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"deployment_type": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Cluster topology: `standalone`, `replicaset`, `sharded`, or `cluster`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
				Validators: []validator.String{
					stringvalidator.OneOf(
						client.DeploymentStandalone,
						client.DeploymentReplicaSet,
						client.DeploymentSharded,
						client.DeploymentCluster,
					),
				},
			},
			"cloud_profile_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID of the cloud profile that determines the cloud provider and credentials used to provision the cluster.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"region": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Cloud region in which to deploy the cluster (for example `us-east-1`).",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"size_id": schema.StringAttribute{
				Required:    true,
				Description: "Instance size identifier. Changing this scales the cluster in place.",
			},
			"disk_size_gb": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				Description:   "Disk size in GB. Can be increased in place; decreases force replacement.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"shard_count": schema.Int64Attribute{
				Optional:      true,
				Description:   "Number of shards (only applies to sharded deployments).",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"ssl_enabled": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Default:       booldefault.StaticBool(true),
				Description:   "Whether SSL/TLS is enabled for client connections.",
				PlanModifiers: []planmodifier.Bool{boolRequiresReplace()},
			},
			"encryption_at_rest": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Default:       booldefault.StaticBool(false),
				Description:   "Whether encryption at rest is enabled.",
				PlanModifiers: []planmodifier.Bool{boolRequiresReplace()},
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Optional list of tags applied to the cluster.",
			},

			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current lifecycle status of the cluster.",
			},
			"host": schema.StringAttribute{
				Computed:    true,
				Description: "Primary hostname for connecting to the cluster.",
			},
			"port": schema.Int64Attribute{
				Computed:    true,
				Description: "Port for connecting to the cluster.",
			},
			"connection_string": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "Connection string for the cluster.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp at which the cluster was created.",
			},
		},
	}
}

func (r *clusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, err := clientFromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected provider data", err.Error())
		return
	}
	r.client = c
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, diags := tagsFromList(ctx, plan.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.ClusterCreateRequest{
		Name:             plan.Name.ValueString(),
		DatabaseType:     plan.DatabaseType.ValueString(),
		Version:          stringValue(plan.Version),
		DeploymentType:   stringValue(plan.DeploymentType),
		CloudProfileID:   plan.CloudProfileID.ValueString(),
		Region:           stringValue(plan.Region),
		SizeID:           plan.SizeID.ValueString(),
		DiskSizeGB:       plan.DiskSizeGB.ValueInt64(),
		ShardCount:       plan.ShardCount.ValueInt64(),
		SSLEnabled:       plan.SSLEnabled.ValueBool(),
		EncryptionAtRest: plan.EncryptionAtRest.ValueBool(),
		Tags:             tags,
	}

	cluster, err := r.client.CreateCluster(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating cluster", err.Error())
		return
	}

	tflog.Info(ctx, "created scalegrid cluster, waiting for it to become ready", map[string]any{"id": cluster.ID})

	ready, err := r.client.WaitForClusterReady(ctx, cluster.ID, clusterPollInterval)
	if err != nil {
		resp.Diagnostics.AddError("Error waiting for cluster to become ready", err.Error())
		// Persist the ID so a subsequent apply or destroy can reconcile.
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), cluster.ID)...)
		return
	}

	resp.Diagnostics.Append(r.mapClusterToState(ctx, ready, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cluster, err := r.client.GetCluster(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading cluster", err.Error())
		return
	}

	resp.Diagnostics.Append(r.mapClusterToState(ctx, cluster, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan clusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags, diags := tagsFromList(ctx, plan.Tags)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.ClusterUpdateRequest{
		Name:       plan.Name.ValueString(),
		SizeID:     plan.SizeID.ValueString(),
		DiskSizeGB: plan.DiskSizeGB.ValueInt64(),
		Tags:       tags,
	}

	if _, err := r.client.UpdateCluster(ctx, plan.ID.ValueString(), updateReq); err != nil {
		resp.Diagnostics.AddError("Error updating cluster", err.Error())
		return
	}

	// Scaling/disk changes are applied asynchronously; wait for the cluster to
	// settle so computed attributes reflect reality.
	cluster, err := r.client.WaitForClusterReady(ctx, plan.ID.ValueString(), clusterPollInterval)
	if err != nil {
		resp.Diagnostics.AddError("Error waiting for cluster update", err.Error())
		return
	}

	resp.Diagnostics.Append(r.mapClusterToState(ctx, cluster, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCluster(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting cluster", err.Error())
		return
	}

	if err := r.client.WaitForClusterDeleted(ctx, state.ID.ValueString(), clusterPollInterval); err != nil {
		resp.Diagnostics.AddError("Error waiting for cluster deletion", err.Error())
		return
	}
}

func (r *clusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapClusterToState copies API data into the Terraform model, preserving plan
// values for fields the API does not echo back.
func (r *clusterResource) mapClusterToState(ctx context.Context, cluster *client.Cluster, model *clusterResourceModel) diagnosticsList {
	var diags diagnosticsList

	model.ID = types.StringValue(cluster.ID)
	model.Name = types.StringValue(cluster.Name)
	model.DatabaseType = types.StringValue(cluster.DatabaseType)
	model.CloudProfileID = optionalString(cluster.CloudProfileID)
	model.SizeID = types.StringValue(cluster.SizeID)
	model.SSLEnabled = types.BoolValue(cluster.SSLEnabled)
	model.EncryptionAtRest = types.BoolValue(cluster.EncryptionAtRest)

	if cluster.Version != "" {
		model.Version = types.StringValue(cluster.Version)
	}
	if cluster.DeploymentType != "" {
		model.DeploymentType = types.StringValue(cluster.DeploymentType)
	}
	if cluster.Region != "" {
		model.Region = types.StringValue(cluster.Region)
	}
	if cluster.DiskSizeGB > 0 {
		model.DiskSizeGB = types.Int64Value(cluster.DiskSizeGB)
	}
	if cluster.ShardCount > 0 {
		model.ShardCount = types.Int64Value(cluster.ShardCount)
	}

	model.Status = optionalString(cluster.Status)
	model.Host = optionalString(cluster.Host)
	model.Port = types.Int64Value(cluster.Port)
	model.ConnectionString = optionalString(cluster.ConnectionString)
	model.CreatedAt = optionalString(cluster.CreatedAt)

	if len(cluster.Tags) > 0 {
		list, d := tagsToList(ctx, cluster.Tags)
		diags = append(diags, d...)
		model.Tags = list
	}

	return diags
}

// boolRequiresReplace returns a plan modifier that forces replacement when a
// boolean attribute changes. It wraps the generic RequiresReplace helper.
func boolRequiresReplace() planmodifier.Bool {
	return boolReplace{}
}

type boolReplace struct{}

func (boolReplace) Description(context.Context) string {
	return "Changing this value forces a new cluster."
}
func (boolReplace) MarkdownDescription(context.Context) string {
	return "Changing this value forces a new cluster."
}
func (boolReplace) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	// No-op on create/destroy; only replace when a known prior value changes.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}
	if !req.StateValue.Equal(req.PlanValue) {
		resp.RequiresReplace = true
	}
}
