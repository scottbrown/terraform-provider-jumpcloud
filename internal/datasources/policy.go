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

var _ datasource.DataSource = &PolicyDataSource{}

type PolicyDataSource struct {
	client *client.Client
}

type PolicyDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	TemplateID types.String `tfsdk:"template_id"`
	State      types.String `tfsdk:"state"`
	Notes      types.String `tfsdk:"notes"`
	Values     types.String `tfsdk:"values"`
}

func NewPolicyDataSource() datasource.DataSource {
	return &PolicyDataSource{}
}

func (d *PolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (d *PolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the policy.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the policy.",
				Computed:    true,
			},
			"template_id": schema.StringAttribute{
				Description: "The ID of the policy template.",
				Computed:    true,
			},
			"state": schema.StringAttribute{
				Description: "The state of the policy.",
				Computed:    true,
			},
			"notes": schema.StringAttribute{
				Description: "Notes or description for the policy.",
				Computed:    true,
			},
			"values": schema.StringAttribute{
				Description: "JSON-encoded policy values/settings.",
				Computed:    true,
			},
		},
	}
}

func (d *PolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.GetPolicy(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading policy", err.Error())
		return
	}

	config.Name = types.StringValue(policy.Name)
	if policy.TemplateID != "" {
		config.TemplateID = types.StringValue(policy.TemplateID)
	} else {
		config.TemplateID = types.StringNull()
	}
	if policy.State != "" {
		config.State = types.StringValue(policy.State)
	} else {
		config.State = types.StringNull()
	}
	if policy.Notes != "" {
		config.Notes = types.StringValue(policy.Notes)
	} else {
		config.Notes = types.StringNull()
	}
	if len(policy.Values) > 0 && string(policy.Values) != "null" {
		config.Values = types.StringValue(string(policy.Values))
	} else {
		config.Values = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
