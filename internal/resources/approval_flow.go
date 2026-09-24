// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &ApprovalFlowResource{}
var _ resource.ResourceWithImportState = &ApprovalFlowResource{}

type ApprovalFlowResource struct {
	client *client.Client
}

type ApprovalFlowResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	ApprovalType        types.String `tfsdk:"approval_type"`
	ApproverRequirement types.String `tfsdk:"approver_requirement"`
	Category            types.String `tfsdk:"category"`
	ResourceID          types.String `tfsdk:"resource_id"`
	ResourceType        types.String `tfsdk:"resource_type"`
	Status              types.String `tfsdk:"status"`
	TimeBasedAccess     types.Bool   `tfsdk:"time_based_access"`
	FixedDuration       types.Bool   `tfsdk:"fixed_duration"`
	NonAdminApproval    types.Bool   `tfsdk:"non_admin_approval"`
	SlackEnabled        types.Bool   `tfsdk:"slack_enabled"`
	VisibleTo           types.List   `tfsdk:"visible_to"`
}

func NewApprovalFlowResource() resource.Resource {
	return &ApprovalFlowResource{}
}

func (r *ApprovalFlowResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_approval_flow"
}

func (r *ApprovalFlowResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud access request approval flow.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the approval flow.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the approval flow.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the approval flow.",
				Optional:    true,
			},
			"approval_type": schema.StringAttribute{
				Description: "The type of approval: manual or auto.",
				Required:    true,
			},
			"approver_requirement": schema.StringAttribute{
				Description: "Approver requirement: at_least_one or all.",
				Optional:    true,
				Computed:    true,
			},
			"category": schema.StringAttribute{
				Description: "The category of the approval flow.",
				Optional:    true,
				Computed:    true,
			},
			"resource_id": schema.StringAttribute{
				Description: "The ID of the resource this flow grants access to (e.g. a user group ID).",
				Required:    true,
			},
			"resource_type": schema.StringAttribute{
				Description: "The type of the resource (e.g. user_group).",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the approval flow: active or inactive.",
				Optional:    true,
			},
			"time_based_access": schema.BoolAttribute{
				Description: "Whether access is time-limited.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"fixed_duration": schema.BoolAttribute{
				Description: "Whether the duration is fixed (vs. user-selectable).",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"non_admin_approval": schema.BoolAttribute{
				Description: "Whether non-admin users can approve requests.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"slack_enabled": schema.BoolAttribute{
				Description: "Whether Slack notifications are enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"visible_to": schema.ListAttribute{
				Description: "List of group IDs that can see and submit this approval flow.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *ApprovalFlowResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApprovalFlowResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApprovalFlowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	flow := r.modelToAPI(&plan)
	created, err := r.client.CreateApprovalFlow(ctx, flow)
	if err != nil {
		resp.Diagnostics.AddError("Error creating approval flow", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApprovalFlowResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApprovalFlowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	flow, err := r.client.GetApprovalFlow(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading approval flow", err.Error())
		return
	}

	r.apiToModel(flow, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApprovalFlowResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApprovalFlowResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ApprovalFlowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	flow := r.modelToAPI(&plan)
	_, err := r.client.UpdateApprovalFlow(ctx, state.ID.ValueString(), flow)
	if err != nil {
		resp.Diagnostics.AddError("Error updating approval flow", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApprovalFlowResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApprovalFlowResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApprovalFlow(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting approval flow", err.Error())
		return
	}
}

func (r *ApprovalFlowResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	flow, err := r.client.GetApprovalFlow(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing approval flow", fmt.Sprintf("Cannot read approval flow %s: %s", req.ID, err))
		return
	}

	var state ApprovalFlowResourceModel
	state.ID = types.StringValue(flow.ID)
	r.apiToModel(flow, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApprovalFlowResource) modelToAPI(model *ApprovalFlowResourceModel) *client.ApprovalFlow {
	flow := &client.ApprovalFlow{
		Name:                model.Name.ValueString(),
		Description:         model.Description.ValueString(),
		ApprovalType:        model.ApprovalType.ValueString(),
		ApproverRequirement: model.ApproverRequirement.ValueString(),
		Category:            model.Category.ValueString(),
		ResourceID:          model.ResourceID.ValueString(),
		ResourceType:        model.ResourceType.ValueString(),
		Status:              model.Status.ValueString(),
		TimeBasedAccess:     model.TimeBasedAccess.ValueBool(),
		NonAdminApproval:    model.NonAdminApproval.ValueBool(),
		SlackEnabled:        model.SlackEnabled.ValueBool(),
	}

	if model.FixedDuration.ValueBool() {
		flow.FixedDuration = "P1D"
	}

	if !model.VisibleTo.IsNull() {
		var visibleToIDs []string
		model.VisibleTo.ElementsAs(context.Background(), &visibleToIDs, false)
		entries := make([]client.VisibleToEntry, len(visibleToIDs))
		for i, id := range visibleToIDs {
			entries[i] = client.VisibleToEntry{ID: id}
		}
		flow.VisibleTo = entries
	}

	return flow
}

func (r *ApprovalFlowResource) apiToModel(flow *client.ApprovalFlow, model *ApprovalFlowResourceModel) {
	model.Name = types.StringValue(flow.Name)
	if flow.Description != "" {
		model.Description = types.StringValue(flow.Description)
	} else {
		model.Description = types.StringNull()
	}
	model.ApprovalType = types.StringValue(flow.ApprovalType)
	if flow.ApproverRequirement != "" {
		model.ApproverRequirement = types.StringValue(flow.ApproverRequirement)
	}
	if flow.Category != "" {
		model.Category = types.StringValue(flow.Category)
	}
	model.ResourceID = types.StringValue(flow.GetResourceID())
	model.ResourceType = types.StringValue(flow.GetResourceType())
	model.Status = types.StringValue(flow.Status)
	model.TimeBasedAccess = types.BoolValue(flow.TimeBasedAccess)
	model.FixedDuration = types.BoolValue(flow.FixedDuration != "")
	model.NonAdminApproval = types.BoolValue(flow.NonAdminApproval)
	model.SlackEnabled = types.BoolValue(flow.SlackEnabled)

	visibleIDs := flow.GetVisibleToIDs()
	if len(visibleIDs) > 0 {
		elems := make([]attr.Value, len(visibleIDs))
		for i, v := range visibleIDs {
			elems[i] = types.StringValue(v)
		}
		model.VisibleTo, _ = types.ListValue(types.StringType, elems)
	} else {
		model.VisibleTo = types.ListNull(types.StringType)
	}
}
