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

var _ resource.Resource = &CommandAssociationResource{}
var _ resource.ResourceWithImportState = &CommandAssociationResource{}

type CommandAssociationResource struct {
	client *client.Client
}

type CommandAssociationResourceModel struct {
	ID         types.String `tfsdk:"id"`
	CommandID  types.String `tfsdk:"command_id"`
	TargetID   types.String `tfsdk:"target_id"`
	TargetType types.String `tfsdk:"target_type"`
}

func NewCommandAssociationResource() resource.Resource {
	return &CommandAssociationResource{}
}

func (r *CommandAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_command_association"
}

func (r *CommandAssociationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Associates a JumpCloud command with a system or system group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite ID in the format command_id/target_type/target_id.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"command_id": schema.StringAttribute{
				Description: "The ID of the command.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_id": schema.StringAttribute{
				Description: "The ID of the target system or system group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_type": schema.StringAttribute{
				Description: "The type of target: system or system_group.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *CommandAssociationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CommandAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CommandAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assoc := &client.CommandAssociation{
		ID:   plan.TargetID.ValueString(),
		Type: plan.TargetType.ValueString(),
	}

	err := r.client.AddCommandAssociation(ctx, plan.CommandID.ValueString(), assoc)
	if err != nil {
		resp.Diagnostics.AddError("Error creating command association", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", plan.CommandID.ValueString(), plan.TargetType.ValueString(), plan.TargetID.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CommandAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CommandAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assocs, err := r.client.GetCommandAssociations(ctx, state.CommandID.ValueString(), state.TargetType.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading command association", err.Error())
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

func (r *CommandAssociationResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Command associations cannot be updated in place. Delete and recreate instead.")
}

func (r *CommandAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CommandAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	assoc := &client.CommandAssociation{
		ID:   state.TargetID.ValueString(),
		Type: state.TargetType.ValueString(),
	}

	err := r.client.RemoveCommandAssociation(ctx, state.CommandID.ValueString(), assoc)
	if err != nil {
		resp.Diagnostics.AddError("Error removing command association", err.Error())
		return
	}
}

func (r *CommandAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected format: command_id/target_type/target_id, got: %s", req.ID))
		return
	}

	state := CommandAssociationResourceModel{
		ID:         types.StringValue(req.ID),
		CommandID:  types.StringValue(parts[0]),
		TargetType: types.StringValue(parts[1]),
		TargetID:   types.StringValue(parts[2]),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
