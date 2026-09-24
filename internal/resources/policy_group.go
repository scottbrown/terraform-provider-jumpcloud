// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &PolicyGroupResource{}
var _ resource.ResourceWithImportState = &PolicyGroupResource{}

type PolicyGroupResource struct {
	client *client.Client
}

type PolicyGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func NewPolicyGroupResource() resource.Resource {
	return &PolicyGroupResource{}
}

func (r *PolicyGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_group"
}

func (r *PolicyGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud policy group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the policy group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the policy group.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the policy group.",
				Optional:    true,
			},
		},
	}
}

func (r *PolicyGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PolicyGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := &client.PolicyGroup{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	created, err := r.client.CreatePolicyGroup(ctx, group)
	if err != nil {
		resp.Diagnostics.AddError("Error creating policy group", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PolicyGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetPolicyGroup(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading policy group", err.Error())
		return
	}

	state.Name = types.StringValue(group.Name)
	if group.Description != "" {
		state.Description = types.StringValue(group.Description)
	} else {
		state.Description = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PolicyGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PolicyGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := &client.PolicyGroup{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	_, err := r.client.UpdatePolicyGroup(ctx, state.ID.ValueString(), group)
	if err != nil {
		resp.Diagnostics.AddError("Error updating policy group", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PolicyGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PolicyGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePolicyGroup(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting policy group", err.Error())
		return
	}
}

func (r *PolicyGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	group, err := r.client.GetPolicyGroup(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing policy group", fmt.Sprintf("Cannot read policy group %s: %s", req.ID, err))
		return
	}

	state := PolicyGroupResourceModel{
		ID:   types.StringValue(group.ID),
		Name: types.StringValue(group.Name),
	}
	if group.Description != "" {
		state.Description = types.StringValue(group.Description)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
