// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ datasource.DataSource = &ApprovalFlowDataSource{}

type ApprovalFlowDataSource struct {
	client *client.Client
}

type ApprovalFlowDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	ApprovalType     types.String `tfsdk:"approval_type"`
	ResourceID       types.String `tfsdk:"resource_id"`
	ResourceType     types.String `tfsdk:"resource_type"`
	Status           types.String `tfsdk:"status"`
	TimeBasedAccess  types.Bool   `tfsdk:"time_based_access"`
	NonAdminApproval types.Bool   `tfsdk:"non_admin_approval"`
}

func NewApprovalFlowDataSource() datasource.DataSource {
	return &ApprovalFlowDataSource{}
}

func (d *ApprovalFlowDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_approval_flow"
}

func (d *ApprovalFlowDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud approval flow.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the approval flow.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the approval flow.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the approval flow.",
				Computed:    true,
			},
			"approval_type": schema.StringAttribute{
				Description: "The approval type.",
				Computed:    true,
			},
			"resource_id": schema.StringAttribute{
				Description: "The resource ID this flow grants access to.",
				Computed:    true,
			},
			"resource_type": schema.StringAttribute{
				Description: "The resource type.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the approval flow.",
				Computed:    true,
			},
			"time_based_access": schema.BoolAttribute{
				Description: "Whether access is time-limited.",
				Computed:    true,
			},
			"non_admin_approval": schema.BoolAttribute{
				Description: "Whether non-admins can approve.",
				Computed:    true,
			},
		},
	}
}

func (d *ApprovalFlowDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", "Expected *client.Client")
		return
	}
	d.client = c
}

func (d *ApprovalFlowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ApprovalFlowDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	flow, err := d.client.GetApprovalFlow(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading approval flow", err.Error())
		return
	}

	config.Name = types.StringValue(flow.Name)
	config.Description = types.StringValue(flow.Description)
	config.ApprovalType = types.StringValue(flow.ApprovalType)
	config.ResourceID = types.StringValue(flow.GetResourceID())
	config.ResourceType = types.StringValue(flow.GetResourceType())
	config.Status = types.StringValue(flow.Status)
	config.TimeBasedAccess = types.BoolValue(flow.TimeBasedAccess)
	config.NonAdminApproval = types.BoolValue(flow.NonAdminApproval)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
