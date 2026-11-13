// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

func TestNotificationChannelResource_Metadata(t *testing.T) {
	r := NewNotificationChannelResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_notification_channel" {
		t.Errorf("expected type name 'jumpcloud_notification_channel', got %q", resp.TypeName)
	}
}

func TestNotificationChannelResource_Schema(t *testing.T) {
	r := NewNotificationChannelResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	required := []string{"name", "type", "configuration"}
	for _, name := range required {
		attr, ok := attrs[name]
		if !ok {
			t.Errorf("expected attribute %q to exist", name)
			continue
		}
		if !attr.IsRequired() {
			t.Errorf("expected attribute %q to be required", name)
		}
	}

	optional := []string{"description", "enabled"}
	for _, name := range optional {
		attr, ok := attrs[name]
		if !ok {
			t.Errorf("expected attribute %q to exist", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("expected attribute %q to be optional", name)
		}
	}

	if attr, ok := attrs["id"]; !ok {
		t.Error("expected attribute 'id' to exist")
	} else if !attr.IsComputed() {
		t.Error("expected attribute 'id' to be computed")
	}
}

func TestNotificationChannelResource_ModelToAPI(t *testing.T) {
	r := &NotificationChannelResource{}

	model := NotificationChannelResourceModel{
		Name:          types.StringValue("Prod Webhook"),
		Description:   types.StringValue("Production alerts"),
		Type:          types.StringValue("CHANNEL_TYPE_WEBHOOK"),
		Enabled:       types.BoolValue(true),
		Configuration: types.StringValue(`{"webhook":{"url":"https://hooks.example.com","sslVerification":true}}`),
	}

	channel := r.modelToAPI(&model)

	if channel.Name != "Prod Webhook" {
		t.Errorf("expected name 'Prod Webhook', got %q", channel.Name)
	}
	if channel.Type != "CHANNEL_TYPE_WEBHOOK" {
		t.Errorf("expected type 'CHANNEL_TYPE_WEBHOOK', got %q", channel.Type)
	}
	if !channel.Enabled {
		t.Error("expected enabled to be true")
	}
	if channel.Description != "Production alerts" {
		t.Errorf("expected description 'Production alerts', got %q", channel.Description)
	}
	if string(channel.Config) != `{"webhook":{"url":"https://hooks.example.com","sslVerification":true}}` {
		t.Errorf("expected config to match, got %q", string(channel.Config))
	}
}

func TestNotificationChannelResource_ModelToAPI_NullOptionals(t *testing.T) {
	r := &NotificationChannelResource{}

	model := NotificationChannelResourceModel{
		Name:          types.StringValue("Minimal"),
		Description:   types.StringNull(),
		Type:          types.StringValue("CHANNEL_TYPE_SLACK"),
		Enabled:       types.BoolValue(false),
		Configuration: types.StringValue(`{"slack":{"slackChannel":[]}}`),
	}

	channel := r.modelToAPI(&model)

	if channel.Description != "" {
		t.Errorf("expected empty description, got %q", channel.Description)
	}
}

func TestNotificationChannelResource_APIToModel(t *testing.T) {
	r := &NotificationChannelResource{}

	t.Run("with config", func(t *testing.T) {
		channel := &client.NotificationChannel{
			ID:          "ch-123",
			Name:        "Slack Channel",
			Description: "Team notifications",
			Type:        "CHANNEL_TYPE_SLACK",
			Enabled:     true,
			Config:      []byte(`{"slack": {"slackChannel": [{"slackChannelId": "C123"}]}}`),
		}

		var model NotificationChannelResourceModel
		r.apiToModel(channel, &model)

		if model.ID.ValueString() != "ch-123" {
			t.Errorf("expected id 'ch-123', got %q", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Slack Channel" {
			t.Errorf("expected name 'Slack Channel', got %q", model.Name.ValueString())
		}
		if model.Type.ValueString() != "CHANNEL_TYPE_SLACK" {
			t.Errorf("expected type 'CHANNEL_TYPE_SLACK', got %q", model.Type.ValueString())
		}
		if !model.Enabled.ValueBool() {
			t.Error("expected enabled to be true")
		}
		if model.Description.ValueString() != "Team notifications" {
			t.Errorf("expected description 'Team notifications', got %q", model.Description.ValueString())
		}
		if model.Configuration.IsNull() {
			t.Error("expected configuration to not be null")
		}
	})

	t.Run("null config", func(t *testing.T) {
		channel := &client.NotificationChannel{
			ID:   "ch-456",
			Name: "Empty",
			Type: "CHANNEL_TYPE_EMAIL",
		}

		var model NotificationChannelResourceModel
		r.apiToModel(channel, &model)

		if model.Configuration.IsNull() != true {
			t.Errorf("expected configuration to be null, got %q", model.Configuration.ValueString())
		}
		if model.Description.IsNull() != true {
			t.Errorf("expected description to be null, got %q", model.Description.ValueString())
		}
	})
}
