package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/davispalomino/terraform-provider-vtex/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Check that types satisfy framework interfaces
var _ resource.Resource = &VtexUserRoleResource{}
var _ resource.ResourceWithImportState = &VtexUserRoleResource{}

func NewVtexUserRoleResource() resource.Resource {
	return &VtexUserRoleResource{}
}

// VtexUserRoleResource is the resource implementation
type VtexUserRoleResource struct {
	client *client.VtexClient
}

// VtexUserRoleResourceModel is the resource data model
type VtexUserRoleResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Email    types.String `tfsdk:"email"`
	Name     types.String `tfsdk:"name"`
	Account  types.String `tfsdk:"account"`
	RoleName types.String `tfsdk:"role_name"`
}

func (r *VtexUserRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_role"
}

func (r *VtexUserRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user with a specific role in a VTEX account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique ID of the resource (email:account:role_name)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "User email",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "User name (if not given, it is taken from email)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account": schema.StringAttribute{
				Required:    true,
				Description: "VTEX account where the role will be assigned (e.g. vendor)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_name": schema.StringAttribute{
				Required:    true,
				Description: "Role name to assign (e.g. Owner, Operation)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *VtexUserRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if provider is not configured
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.VtexClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.VtexClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *VtexUserRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VtexUserRoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If name is not given, get it from email
	name := data.Name.ValueString()
	if name == "" {
		emailParts := strings.Split(data.Email.ValueString(), "@")
		name = emailParts[0]
		data.Name = types.StringValue(name)
	}

	// Create user in VTEX
	userRole := client.UserRole{
		Email:    data.Email.ValueString(),
		Name:     name,
		Account:  data.Account.ValueString(),
		RoleName: data.RoleName.ValueString(),
	}

	tflog.Debug(ctx, "Creating VTEX user role", map[string]interface{}{
		"email":     userRole.Email,
		"account":   userRole.Account,
		"role_name": userRole.RoleName,
	})

	err := r.client.CreateUserRole(userRole)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating VTEX User Role",
			"Could not create user role, unexpected error: "+err.Error(),
		)
		return
	}

	// Generate unique ID
	data.ID = types.StringValue(fmt.Sprintf("%s:%s:%s",
		data.Email.ValueString(),
		data.Account.ValueString(),
		data.RoleName.ValueString(),
	))

	tflog.Trace(ctx, "Created VTEX user role", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VtexUserRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VtexUserRoleResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// VTEX does not have an endpoint to query specific users
	// We assume the resource exists if it is in the state

	tflog.Debug(ctx, "Reading VTEX user role", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VtexUserRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VtexUserRoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Main fields (email, account, role_name) have RequiresReplace
	// Any change will destroy and recreate the resource
	// VTEX does not have an update endpoint, so this is a no-op

	tflog.Debug(ctx, "Update called for VTEX user role (no-op)", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VtexUserRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VtexUserRoleResourceModel

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete user from VTEX
	userRole := client.UserRole{
		Email:    data.Email.ValueString(),
		Name:     data.Name.ValueString(),
		Account:  data.Account.ValueString(),
		RoleName: data.RoleName.ValueString(),
	}

	tflog.Debug(ctx, "Deleting VTEX user role", map[string]interface{}{
		"email":     userRole.Email,
		"account":   userRole.Account,
		"role_name": userRole.RoleName,
	})

	err := r.client.DeleteUserRole(userRole)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting VTEX User Role",
			"Could not delete user role, unexpected error: "+err.Error(),
		)
		return
	}

	tflog.Trace(ctx, "Deleted VTEX user role", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

func (r *VtexUserRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// ID format: email:account:role_name
	parts := strings.SplitN(req.ID, ":", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: email:account:role_name, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("account"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role_name"), parts[2])...)

	// Get name from email
	emailParts := strings.Split(parts[0], "@")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), emailParts[0])...)
}
