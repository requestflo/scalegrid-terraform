package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/requestflo/scalegrid-terraform/internal/client"
)

var (
	_ resource.Resource                = (*backupResource)(nil)
	_ resource.ResourceWithConfigure   = (*backupResource)(nil)
	_ resource.ResourceWithImportState = (*backupResource)(nil)
)

// NewBackupResource is the constructor registered with the provider.
func NewBackupResource() resource.Resource {
	return &backupResource{}
}

type backupResource struct {
	client *client.Client
}

type backupResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ClusterID types.String `tfsdk:"cluster_id"`
	Status    types.String `tfsdk:"status"`
	SizeBytes types.Int64  `tfsdk:"size_bytes"`
	Type      types.String `tfsdk:"type"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *backupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (r *backupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers and manages an on-demand backup of a ScaleGrid cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique identifier of the backup.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"cluster_id": schema.StringAttribute{
				Required:      true,
				Description:   "ID of the cluster to back up.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status of the backup.",
			},
			"size_bytes": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of the backup in bytes.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of the backup (for example `on_demand` or `scheduled`).",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp at which the backup was created.",
			},
		},
	}
}

func (r *backupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, err := clientFromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected provider data", err.Error())
		return
	}
	r.client = c
}

func (r *backupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup, err := r.client.CreateBackup(ctx, plan.ClusterID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating backup", err.Error())
		return
	}

	r.mapToState(backup, plan.ClusterID.ValueString(), &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *backupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state backupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup, err := r.client.GetBackup(ctx, state.ClusterID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading backup", err.Error())
		return
	}

	r.mapToState(backup, state.ClusterID.ValueString(), &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is a no-op: a backup has no mutable attributes. Any change to
// cluster_id forces replacement via the schema plan modifier.
func (r *backupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan backupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *backupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state backupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteBackup(ctx, state.ClusterID.ValueString(), state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting backup", err.Error())
	}
}

// ImportState accepts a "<cluster_id>:<backup_id>" composite identifier.
func (r *backupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the form \"cluster_id:backup_id\", got %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *backupResource) mapToState(backup *client.Backup, clusterID string, model *backupResourceModel) {
	model.ID = types.StringValue(backup.ID)
	model.Status = optionalString(backup.Status)
	model.SizeBytes = types.Int64Value(backup.SizeBytes)
	model.Type = optionalString(backup.Type)
	model.CreatedAt = optionalString(backup.CreatedAt)
	if backup.ClusterID != "" {
		model.ClusterID = types.StringValue(backup.ClusterID)
	} else {
		model.ClusterID = types.StringValue(clusterID)
	}
}
