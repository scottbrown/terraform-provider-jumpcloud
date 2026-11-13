// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &NotificationChannelResource{}
var _ resource.ResourceWithImportState = &NotificationChannelResource{}

type NotificationChannelResource struct {
	client *client.Client
}

type NotificationChannelResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Type          types.String `tfsdk:"type"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Configuration types.String `tfsdk:"configuration"`
}

func NewNotificationChannelResource() resource.Resource {
	return &NotificationChannelResource{}
}

func (r *NotificationChannelResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_channel"
}

func (r *NotificationChannelResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud notification channel (email, Slack, or webhook).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the notification channel.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the notification channel.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the notification channel.",
				Optional:    true,
			},
			"type": schema.StringAttribute{
				Description: "Channel type: CHANNEL_TYPE_EMAIL, CHANNEL_TYPE_SLACK, or CHANNEL_TYPE_WEBHOOK.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the notification channel is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"configuration": schema.StringAttribute{
				Description: "JSON-encoded channel configuration. Structure depends on channel type (email, slack, or webhook config object).",
				Required:    true,
			},
		},
	}
}

func (r *NotificationChannelResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NotificationChannelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NotificationChannelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channel := r.modelToAPI(&plan)
	created, err := r.client.CreateNotificationChannel(ctx, channel)
	if err != nil {
		resp.Diagnostics.AddError("Error creating notification channel", err.Error())
		return
	}

	r.apiToModel(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationChannelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NotificationChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channel, err := r.client.GetNotificationChannel(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading notification channel", err.Error())
		return
	}

	r.apiToModel(channel, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationChannelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NotificationChannelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state NotificationChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	channel := r.modelToAPI(&plan)
	updated, err := r.client.UpdateNotificationChannel(ctx, state.ID.ValueString(), channel)
	if err != nil {
		resp.Diagnostics.AddError("Error updating notification channel", err.Error())
		return
	}

	r.apiToModel(updated, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NotificationChannelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NotificationChannelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNotificationChannel(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting notification channel", err.Error())
		return
	}
}

func (r *NotificationChannelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	channel, err := r.client.GetNotificationChannel(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing notification channel", fmt.Sprintf("Cannot read notification channel %s: %s", req.ID, err))
		return
	}

	var state NotificationChannelResourceModel
	r.apiToModel(channel, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NotificationChannelResource) modelToAPI(model *NotificationChannelResourceModel) *client.NotificationChannel {
	channel := &client.NotificationChannel{
		Name:    model.Name.ValueString(),
		Type:    model.Type.ValueString(),
		Enabled: model.Enabled.ValueBool(),
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		channel.Description = model.Description.ValueString()
	}
	if !model.Configuration.IsNull() && !model.Configuration.IsUnknown() {
		channel.Config = json.RawMessage(model.Configuration.ValueString())
	}
	return channel
}

func (r *NotificationChannelResource) apiToModel(channel *client.NotificationChannel, model *NotificationChannelResourceModel) {
	model.ID = types.StringValue(channel.ID)
	model.Name = types.StringValue(channel.Name)
	model.Type = types.StringValue(channel.Type)
	model.Enabled = types.BoolValue(channel.Enabled)

	if channel.Description != "" {
		model.Description = types.StringValue(channel.Description)
	} else {
		model.Description = types.StringNull()
	}

	if len(channel.Config) > 0 && string(channel.Config) != "null" {
		var normalized any
		if err := json.Unmarshal(channel.Config, &normalized); err == nil {
			compact, _ := json.Marshal(normalized)
			model.Configuration = types.StringValue(string(compact))
		} else {
			model.Configuration = types.StringValue(string(channel.Config))
		}
	} else {
		model.Configuration = types.StringNull()
	}
}
