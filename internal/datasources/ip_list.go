// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ datasource.DataSource = &IPListDataSource{}

type IPListDataSource struct {
	client *client.Client
}

type IPListDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IPs         types.List   `tfsdk:"ips"`
}

func NewIPListDataSource() datasource.DataSource {
	return &IPListDataSource{}
}

func (d *IPListDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_list"
}

func (d *IPListDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud IP list.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the IP list.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the IP list.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the IP list.",
				Computed:    true,
			},
			"ips": schema.ListAttribute{
				Description: "List of IP addresses or CIDR ranges.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *IPListDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IPListDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config IPListDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipList, err := d.client.GetIPList(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP list", err.Error())
		return
	}

	config.Name = types.StringValue(ipList.Name)
	if ipList.Description != "" {
		config.Description = types.StringValue(ipList.Description)
	} else {
		config.Description = types.StringNull()
	}

	if len(ipList.IPs) > 0 {
		ipValues := make([]attr.Value, len(ipList.IPs))
		for i, ip := range ipList.IPs {
			ipValues[i] = types.StringValue(ip)
		}
		config.IPs, _ = types.ListValue(types.StringType, ipValues)
	} else {
		config.IPs, _ = types.ListValue(types.StringType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
