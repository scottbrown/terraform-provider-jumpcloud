// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ datasource.DataSource = &PolicyTemplateDataSource{}

type PolicyTemplateDataSource struct {
	client *client.Client
}

type PolicyTemplateDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	OSFamily    types.String `tfsdk:"os_family"`
}

func NewPolicyTemplateDataSource() datasource.DataSource {
	return &PolicyTemplateDataSource{}
}

func (d *PolicyTemplateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_template"
}

func (d *PolicyTemplateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud policy template.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the policy template.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the policy template.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the policy template.",
				Computed:    true,
			},
			"os_family": schema.StringAttribute{
				Description: "The OS family this template applies to (e.g. darwin, linux, windows).",
				Computed:    true,
			},
		},
	}
}

func (d *PolicyTemplateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PolicyTemplateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PolicyTemplateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := d.client.GetPolicyTemplate(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading policy template", err.Error())
		return
	}

	config.Name = types.StringValue(tmpl.Name)
	if tmpl.Description != "" {
		config.Description = types.StringValue(tmpl.Description)
	} else {
		config.Description = types.StringNull()
	}
	if tmpl.OSFamily != "" {
		config.OSFamily = types.StringValue(tmpl.OSFamily)
	} else {
		config.OSFamily = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
