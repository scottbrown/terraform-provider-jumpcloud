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

var _ datasource.DataSource = &CommandDataSource{}

type CommandDataSource struct {
	client *client.Client
}

type CommandDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Command     types.String `tfsdk:"command"`
	CommandType types.String `tfsdk:"command_type"`
	User        types.String `tfsdk:"user"`
	Sudo        types.Bool   `tfsdk:"sudo"`
	Shell       types.String `tfsdk:"shell"`
	Trigger     types.String `tfsdk:"trigger"`
}

func NewCommandDataSource() datasource.DataSource {
	return &CommandDataSource{}
}

func (d *CommandDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_command"
}

func (d *CommandDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud command.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the command.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the command.",
				Computed:    true,
			},
			"command": schema.StringAttribute{
				Description: "The script content.",
				Computed:    true,
			},
			"command_type": schema.StringAttribute{
				Description: "The platform type.",
				Computed:    true,
			},
			"user": schema.StringAttribute{
				Description: "Execution user context.",
				Computed:    true,
			},
			"sudo": schema.BoolAttribute{
				Description: "Whether elevated privileges are used.",
				Computed:    true,
			},
			"shell": schema.StringAttribute{
				Description: "Shell used for execution.",
				Computed:    true,
			},
			"trigger": schema.StringAttribute{
				Description: "Trigger type.",
				Computed:    true,
			},
		},
	}
}

func (d *CommandDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CommandDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CommandDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmd, err := d.client.GetCommand(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading command", err.Error())
		return
	}

	config.Name = types.StringValue(cmd.Name)
	config.Command = types.StringValue(cmd.Command)
	config.CommandType = types.StringValue(cmd.CommandType)
	config.User = types.StringValue(cmd.User)
	config.Sudo = types.BoolValue(cmd.Sudo)
	if cmd.Shell != "" {
		config.Shell = types.StringValue(cmd.Shell)
	}
	if cmd.Trigger != "" {
		config.Trigger = types.StringValue(cmd.Trigger)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
