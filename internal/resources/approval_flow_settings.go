// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &ApprovalFlowSettingsResource{}

type ApprovalFlowSettingsResource struct {
	client *client.Client
}

type ApprovalFlowSettingsResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	AdminApprovalEmail     types.Bool   `tfsdk:"admin_approval_email"`
	ExposeApprovalProgress types.Bool   `tfsdk:"expose_approval_progress"`
	ResourceRequest        types.Bool   `tfsdk:"resource_request"`
	UserApprovalEmail      types.Bool   `tfsdk:"user_approval_email"`
	UserRequestEmail       types.Bool   `tfsdk:"user_request_email"`
}

func NewApprovalFlowSettingsResource() resource.Resource {
	return &ApprovalFlowSettingsResource{}
}

func (r *ApprovalFlowSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_approval_flow_settings"
}

func (r *ApprovalFlowSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages organization-level JumpCloud approval flow settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the settings.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"admin_approval_email": schema.BoolAttribute{
				Description: "Whether to send email notifications to admins for approval requests.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"expose_approval_progress": schema.BoolAttribute{
				Description: "Whether to expose approval progress to requesters.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"resource_request": schema.BoolAttribute{
				Description: "Whether resource requests are enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"user_approval_email": schema.BoolAttribute{
				Description: "Whether to send email to users when their request is approved.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"user_request_email": schema.BoolAttribute{
				Description: "Whether to send email confirmation to users when they submit a request.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

func (r *ApprovalFlowSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApprovalFlowSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApprovalFlowSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := r.modelToAPI(&plan)
	created, err := r.client.CreateApprovalFlowSettings(ctx, settings)
	if err != nil {
		resp.Diagnostics.AddError("Error creating approval flow settings", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApprovalFlowSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApprovalFlowSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.GetApprovalFlowSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading approval flow settings", err.Error())
		return
	}

	r.apiToModel(settings, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApprovalFlowSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApprovalFlowSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ApprovalFlowSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := r.modelToAPI(&plan)
	_, err := r.client.UpdateApprovalFlowSettings(ctx, settings)
	if err != nil {
		resp.Diagnostics.AddError("Error updating approval flow settings", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApprovalFlowSettingsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Approval flow settings cannot be deleted",
		"JumpCloud does not support deleting org-level approval flow settings. The resource will be removed from state but settings will remain in JumpCloud.",
	)
}

func (r *ApprovalFlowSettingsResource) modelToAPI(model *ApprovalFlowSettingsResourceModel) *client.ApprovalFlowSettings {
	return &client.ApprovalFlowSettings{
		AdminApprovalEmail:     model.AdminApprovalEmail.ValueBool(),
		ExposeApprovalProgress: model.ExposeApprovalProgress.ValueBool(),
		ResourceRequest:        model.ResourceRequest.ValueBool(),
		UserApprovalEmail:      model.UserApprovalEmail.ValueBool(),
		UserRequestEmail:       model.UserRequestEmail.ValueBool(),
	}
}

func (r *ApprovalFlowSettingsResource) apiToModel(settings *client.ApprovalFlowSettings, model *ApprovalFlowSettingsResourceModel) {
	model.ID = types.StringValue(settings.ID)
	model.AdminApprovalEmail = types.BoolValue(settings.AdminApprovalEmail)
	model.ExposeApprovalProgress = types.BoolValue(settings.ExposeApprovalProgress)
	model.ResourceRequest = types.BoolValue(settings.ResourceRequest)
	model.UserApprovalEmail = types.BoolValue(settings.UserApprovalEmail)
	model.UserRequestEmail = types.BoolValue(settings.UserRequestEmail)
}
