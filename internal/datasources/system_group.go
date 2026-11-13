// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ datasource.DataSource = &SystemGroupDataSource{}

type SystemGroupDataSource struct {
	client *client.Client
}

type SystemGroupDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	MembershipMethod types.String `tfsdk:"membership_method"`
	MemberQuery      types.String `tfsdk:"member_query"`
}

func NewSystemGroupDataSource() datasource.DataSource {
	return &SystemGroupDataSource{}
}

func (d *SystemGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_system_group"
}

func (d *SystemGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud system group by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the system group. Exactly one of id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the system group. Exactly one of id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the system group.",
				Computed:    true,
			},
			"membership_method": schema.StringAttribute{
				Description: "Membership method: STATIC, DYNAMIC_REVIEW_REQUIRED, or DYNAMIC_AUTOMATED.",
				Computed:    true,
			},
			"member_query": schema.StringAttribute{
				Description: "JSON-encoded dynamic membership query.",
				Computed:    true,
			},
		},
	}
}

func (d *SystemGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SystemGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SystemGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && !config.ID.IsUnknown()
	hasName := !config.Name.IsNull() && !config.Name.IsUnknown()

	if !hasID && !hasName {
		resp.Diagnostics.AddError("Missing attribute", "Exactly one of 'id' or 'name' must be specified.")
		return
	}
	if hasID && hasName {
		resp.Diagnostics.AddError("Conflicting attributes", "Only one of 'id' or 'name' may be specified.")
		return
	}

	var group *client.SystemGroup
	var err error

	if hasID {
		group, err = d.client.GetSystemGroup(ctx, config.ID.ValueString())
	} else {
		group, err = d.client.GetSystemGroupByName(ctx, config.Name.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Error reading system group", err.Error())
		return
	}

	config.ID = types.StringValue(group.ID)
	config.Name = types.StringValue(group.Name)

	if group.Description != "" {
		config.Description = types.StringValue(group.Description)
	} else {
		config.Description = types.StringNull()
	}

	if group.MembershipMethod != "" {
		config.MembershipMethod = types.StringValue(group.MembershipMethod)
	} else {
		config.MembershipMethod = types.StringNull()
	}

	if group.MemberQuery != nil && len(group.MemberQuery.Filters) > 0 {
		mqJSON, err := json.Marshal(group.MemberQuery)
		if err == nil {
			config.MemberQuery = types.StringValue(string(mqJSON))
		} else {
			config.MemberQuery = types.StringNull()
		}
	} else {
		config.MemberQuery = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
