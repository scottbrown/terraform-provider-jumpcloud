// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &ApplicationAssociationResource{}
var _ resource.ResourceWithImportState = &ApplicationAssociationResource{}

type ApplicationAssociationResource struct {
	client *client.Client
}

type ApplicationAssociationResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ApplicationID types.String `tfsdk:"application_id"`
	TargetID      types.String `tfsdk:"target_id"`
	TargetType    types.String `tfsdk:"target_type"`
}

func NewApplicationAssociationResource() resource.Resource {
	return &ApplicationAssociationResource{}
}

func (r *ApplicationAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_association"
}

func (r *ApplicationAssociationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Associates a JumpCloud application with a user or user group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID in the format application_id/target_type/target_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_id": schema.StringAttribute{
				Description: "The ID of the target user or user group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_type": schema.StringAttribute{
				Description: "The type of target: user or user_group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ApplicationAssociationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *client.Client")
		return
	}
	r.client = c
}

func (r *ApplicationAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assoc := &client.ApplicationAssociation{
		ID:   plan.TargetID.ValueString(),
		Type: plan.TargetType.ValueString(),
	}

	err := r.client.AddApplicationAssociation(ctx, plan.ApplicationID.ValueString(), assoc)
	if err != nil {
		resp.Diagnostics.AddError("Error creating application association", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", plan.ApplicationID.ValueString(), plan.TargetType.ValueString(), plan.TargetID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assocs, err := r.client.GetApplicationAssociations(ctx, state.ApplicationID.ValueString(), state.TargetType.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading application association", err.Error())
		return
	}

	found := false
	for _, a := range assocs {
		if a.ID == state.TargetID.ValueString() {
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApplicationAssociationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Application associations cannot be updated in place. Delete and recreate instead.")
}

func (r *ApplicationAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assoc := &client.ApplicationAssociation{
		ID:   state.TargetID.ValueString(),
		Type: state.TargetType.ValueString(),
	}

	err := r.client.RemoveApplicationAssociation(ctx, state.ApplicationID.ValueString(), assoc)
	if err != nil {
		resp.Diagnostics.AddError("Error removing application association", err.Error())
		return
	}
}

func (r *ApplicationAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format: application_id/target_type/target_id, got: %s", req.ID))
		return
	}

	state := ApplicationAssociationResourceModel{
		ID:            types.StringValue(req.ID),
		ApplicationID: types.StringValue(parts[0]),
		TargetType:    types.StringValue(parts[1]),
		TargetID:      types.StringValue(parts[2]),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
