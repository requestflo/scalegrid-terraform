package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/requestflo/scalegrid-terraform/internal/client"
)

var (
	_ resource.Resource                = (*cloudProfileResource)(nil)
	_ resource.ResourceWithConfigure   = (*cloudProfileResource)(nil)
	_ resource.ResourceWithImportState = (*cloudProfileResource)(nil)
)

// NewCloudProfileResource is the constructor registered with the provider.
func NewCloudProfileResource() resource.Resource {
	return &cloudProfileResource{}
}

type cloudProfileResource struct {
	client *client.Client
}

type cloudProfileResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	CloudProvider  types.String `tfsdk:"cloud_provider"`
	Region         types.String `tfsdk:"region"`
	AccessKey      types.String `tfsdk:"access_key"`
	SecretKey      types.String `tfsdk:"secret_key"`
	SubscriptionID types.String `tfsdk:"subscription_id"`
	TenantID       types.String `tfsdk:"tenant_id"`
	ClientID       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (r *cloudProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloud_profile"
}

func (r *cloudProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a ScaleGrid cloud profile, which stores the cloud credentials used to " +
			"provision clusters in a Bring Your Own Cloud account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "Unique identifier of the cloud profile.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Human-readable name of the cloud profile.",
			},
			"cloud_provider": schema.StringAttribute{
				Required:      true,
				Description:   "Cloud provider: `aws`, `azure`, `gcp`, `digitalocean`, `oci`, or `vmware`.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.OneOf("aws", "azure", "gcp", "digitalocean", "oci", "vmware"),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Description: "Default cloud region for the profile.",
			},
			"access_key": schema.StringAttribute{
				Optional:    true,
				Description: "Access key (AWS) or equivalent identifier.",
			},
			"secret_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Secret key (AWS) or equivalent credential. Write-only; not returned on read.",
			},
			"subscription_id": schema.StringAttribute{
				Optional:    true,
				Description: "Subscription ID (Azure).",
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Description: "Tenant ID (Azure).",
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "Client/application ID (Azure).",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client secret (Azure). Write-only; not returned on read.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp at which the cloud profile was created.",
			},
		},
	}
}

func (r *cloudProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, err := clientFromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected provider data", err.Error())
		return
	}
	r.client = c
}

func (r *cloudProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cloudProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profile, err := r.client.CreateCloudProfile(ctx, r.toAPI(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating cloud profile", err.Error())
		return
	}

	r.mapToState(profile, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cloudProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profile, err := r.client.GetCloudProfile(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading cloud profile", err.Error())
		return
	}

	r.mapToState(profile, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cloudProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profile, err := r.client.UpdateCloudProfile(ctx, plan.ID.ValueString(), r.toAPI(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating cloud profile", err.Error())
		return
	}

	r.mapToState(profile, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cloudProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state cloudProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCloudProfile(ctx, state.ID.ValueString()); err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting cloud profile", err.Error())
	}
}

func (r *cloudProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *cloudProfileResource) toAPI(model cloudProfileResourceModel) client.CloudProfile {
	return client.CloudProfile{
		Name:           model.Name.ValueString(),
		CloudProvider:  model.CloudProvider.ValueString(),
		Region:         stringValue(model.Region),
		AccessKey:      stringValue(model.AccessKey),
		SecretKey:      stringValue(model.SecretKey),
		SubscriptionID: stringValue(model.SubscriptionID),
		TenantID:       stringValue(model.TenantID),
		ClientID:       stringValue(model.ClientID),
		ClientSecret:   stringValue(model.ClientSecret),
	}
}

// mapToState copies API data into the model. Sensitive write-only fields
// (secret_key, client_secret) are preserved from configuration because the API
// does not return them.
func (r *cloudProfileResource) mapToState(profile *client.CloudProfile, model *cloudProfileResourceModel) {
	model.ID = types.StringValue(profile.ID)
	model.Name = types.StringValue(profile.Name)
	model.CloudProvider = types.StringValue(profile.CloudProvider)
	model.Region = optionalString(profile.Region)
	model.AccessKey = optionalString(profile.AccessKey)
	model.SubscriptionID = optionalString(profile.SubscriptionID)
	model.TenantID = optionalString(profile.TenantID)
	model.ClientID = optionalString(profile.ClientID)
	model.CreatedAt = optionalString(profile.CreatedAt)
	// secret_key and client_secret intentionally left as-is.
}
