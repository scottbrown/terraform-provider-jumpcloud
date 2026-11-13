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

var _ resource.Resource = &PolicyResource{}
var _ resource.ResourceWithImportState = &PolicyResource{}

type PolicyResource struct {
	client *client.Client
}

type PolicyResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	TemplateID types.String `tfsdk:"template_id"`
	State      types.String `tfsdk:"state"`
	Notes      types.String `tfsdk:"notes"`
	Values     types.String `tfsdk:"values"`
}

func NewPolicyResource() resource.Resource {
	return &PolicyResource{}
}

func (r *PolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (r *PolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud policy (MDM configuration profile or device policy).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the policy.",
				Required:    true,
			},
			"template_id": schema.StringAttribute{
				Description: "The ID of the policy template to base this policy on.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Description: "The state of the policy: active or disabled.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"notes": schema.StringAttribute{
				Description: "Notes or description for the policy.",
				Optional:    true,
			},
			"values": schema.StringAttribute{
				Description: "JSON-encoded policy values/settings. Structure depends on the policy template.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *PolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := r.modelToAPI(&plan)
	created, err := r.client.CreatePolicy(ctx, policy)
	if err != nil {
		resp.Diagnostics.AddError("Error creating policy", err.Error())
		return
	}

	r.apiToModel(created, &plan)
	plan.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetPolicy(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading policy", err.Error())
		return
	}

	r.apiToModel(policy, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := r.modelToAPI(&plan)
	updated, err := r.client.UpdatePolicy(ctx, state.ID.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError("Error updating policy", err.Error())
		return
	}

	r.apiToModel(updated, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePolicy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting policy", err.Error())
		return
	}
}

func (r *PolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policy, err := r.client.GetPolicy(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing policy", fmt.Sprintf("Cannot read policy %s: %s", req.ID, err))
		return
	}

	var state PolicyResourceModel
	state.ID = types.StringValue(policy.ID)
	r.apiToModel(policy, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PolicyResource) modelToAPI(model *PolicyResourceModel) *client.Policy {
	policy := &client.Policy{
		Name: model.Name.ValueString(),
	}
	if !model.TemplateID.IsNull() && !model.TemplateID.IsUnknown() {
		policy.TemplateID = model.TemplateID.ValueString()
	}
	if !model.State.IsNull() && !model.State.IsUnknown() {
		policy.State = model.State.ValueString()
	}
	if !model.Notes.IsNull() && !model.Notes.IsUnknown() {
		policy.Notes = model.Notes.ValueString()
	}
	if !model.Values.IsNull() && !model.Values.IsUnknown() {
		policy.Values = json.RawMessage(model.Values.ValueString())
	}
	return policy
}

func (r *PolicyResource) apiToModel(policy *client.Policy, model *PolicyResourceModel) {
	model.Name = types.StringValue(policy.Name)
	if policy.TemplateID != "" {
		model.TemplateID = types.StringValue(policy.TemplateID)
	}
	if policy.State != "" {
		model.State = types.StringValue(policy.State)
	} else {
		model.State = types.StringNull()
	}
	if policy.Notes != "" {
		model.Notes = types.StringValue(policy.Notes)
	} else if model.Notes.IsNull() {
		// preserve null if API returns empty and plan had null
	} else {
		model.Notes = types.StringNull()
	}
	if len(policy.Values) > 0 && string(policy.Values) != "null" {
		var normalized any
		if err := json.Unmarshal(policy.Values, &normalized); err == nil {
			compact, _ := json.Marshal(normalized)
			model.Values = types.StringValue(string(compact))
		} else {
			model.Values = types.StringValue(string(policy.Values))
		}
	} else {
		model.Values = types.StringNull()
	}
}
