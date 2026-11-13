// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &SystemGroupResource{}
var _ resource.ResourceWithImportState = &SystemGroupResource{}

type SystemGroupResource struct {
	client *client.Client
}

type SystemGroupResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	MembershipMethod types.String `tfsdk:"membership_method"`
	MemberQuery      types.String `tfsdk:"member_query"`
}

func NewSystemGroupResource() resource.Resource {
	return &SystemGroupResource{}
}

func (r *SystemGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_group"
}

func (r *SystemGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud system (device) group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the system group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the system group.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the system group.",
				Optional:    true,
			},
			"membership_method": schema.StringAttribute{
				Description: "Membership method: STATIC, DYNAMIC_REVIEW_REQUIRED, or DYNAMIC_AUTOMATED.",
				Optional:    true,
				Computed:    true,
			},
			"member_query": schema.StringAttribute{
				Description: "JSON-encoded dynamic membership query. Required when membership_method is DYNAMIC_REVIEW_REQUIRED or DYNAMIC_AUTOMATED.",
				Optional:    true,
			},
		},
	}
}

func (r *SystemGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SystemGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SystemGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := r.modelToAPI(&plan)
	created, err := r.client.CreateSystemGroup(ctx, group)
	if err != nil {
		resp.Diagnostics.AddError("Error creating system group", err.Error())
		return
	}

	r.apiToModel(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SystemGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SystemGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := r.client.GetSystemGroup(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading system group", err.Error())
		return
	}

	r.apiToModel(group, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SystemGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SystemGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SystemGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := r.modelToAPI(&plan)
	updated, err := r.client.UpdateSystemGroup(ctx, state.ID.ValueString(), group)
	if err != nil {
		resp.Diagnostics.AddError("Error updating system group", err.Error())
		return
	}

	r.apiToModel(updated, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SystemGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SystemGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSystemGroup(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting system group", err.Error())
		return
	}
}

func (r *SystemGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	group, err := r.client.GetSystemGroup(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing system group", fmt.Sprintf("Cannot read system group %s: %s", req.ID, err))
		return
	}

	var state SystemGroupResourceModel
	r.apiToModel(group, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SystemGroupResource) modelToAPI(model *SystemGroupResourceModel) *client.SystemGroup {
	group := &client.SystemGroup{
		Name: model.Name.ValueString(),
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		group.Description = model.Description.ValueString()
	}
	if !model.MembershipMethod.IsNull() && !model.MembershipMethod.IsUnknown() {
		group.MembershipMethod = model.MembershipMethod.ValueString()
	}
	if !model.MemberQuery.IsNull() && !model.MemberQuery.IsUnknown() {
		var mq client.SystemGroupMemberQuery
		if err := json.Unmarshal([]byte(model.MemberQuery.ValueString()), &mq); err == nil {
			group.MemberQuery = &mq
		}
	}
	return group
}

func (r *SystemGroupResource) apiToModel(group *client.SystemGroup, model *SystemGroupResourceModel) {
	model.ID = types.StringValue(group.ID)
	model.Name = types.StringValue(group.Name)

	if group.Description != "" {
		model.Description = types.StringValue(group.Description)
	} else {
		model.Description = types.StringNull()
	}

	if group.MembershipMethod != "" {
		model.MembershipMethod = types.StringValue(group.MembershipMethod)
	} else {
		model.MembershipMethod = types.StringNull()
	}

	if group.MemberQuery != nil && (len(group.MemberQuery.Filters) > 0 || group.MemberQuery.SearchFilters != "") {
		mqJSON, err := json.Marshal(group.MemberQuery)
		if err == nil {
			model.MemberQuery = types.StringValue(string(mqJSON))
		}
	} else {
		model.MemberQuery = types.StringNull()
	}
}
