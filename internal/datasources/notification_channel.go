// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package datasources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ datasource.DataSource = &NotificationChannelDataSource{}

type NotificationChannelDataSource struct {
	client *client.Client
}

type NotificationChannelDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Type          types.String `tfsdk:"type"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Configuration types.String `tfsdk:"configuration"`
}

func NewNotificationChannelDataSource() datasource.DataSource {
	return &NotificationChannelDataSource{}
}

func (d *NotificationChannelDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (d *NotificationChannelDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud notification channel by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the notification channel. Exactly one of id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the notification channel. Exactly one of id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the notification channel.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "Channel type: CHANNEL_TYPE_EMAIL, CHANNEL_TYPE_SLACK, or CHANNEL_TYPE_WEBHOOK.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the notification channel is active.",
				Computed:    true,
			},
			"configuration": schema.StringAttribute{
				Description: "JSON-encoded channel configuration.",
				Computed:    true,
			},
		},
	}
}

func (d *NotificationChannelDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NotificationChannelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NotificationChannelDataSourceModel
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

	var channel *client.NotificationChannel
	var err error

	if hasID {
		channel, err = d.client.GetNotificationChannel(ctx, config.ID.ValueString())
	} else {
		channel, err = d.client.GetNotificationChannelByName(ctx, config.Name.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Error reading notification channel", err.Error())
		return
	}

	config.ID = types.StringValue(channel.ID)
	config.Name = types.StringValue(channel.Name)
	config.Type = types.StringValue(channel.Type)
	config.Enabled = types.BoolValue(channel.Enabled)

	if channel.Description != "" {
		config.Description = types.StringValue(channel.Description)
	} else {
		config.Description = types.StringNull()
	}

	if len(channel.Config) > 0 && string(channel.Config) != "null" {
		var normalized any
		if err := json.Unmarshal(channel.Config, &normalized); err == nil {
			compact, _ := json.Marshal(normalized)
			config.Configuration = types.StringValue(string(compact))
		} else {
			config.Configuration = types.StringValue(string(channel.Config))
		}
	} else {
		config.Configuration = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
