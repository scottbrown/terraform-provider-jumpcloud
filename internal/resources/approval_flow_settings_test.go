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

func TestApprovalFlowSettingsResource_Metadata(t *testing.T) {
	r := NewApprovalFlowSettingsResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_approval_flow_settings" {
		t.Errorf("expected type name 'jumpcloud_approval_flow_settings', got %q", resp.TypeName)
	}
}

func TestApprovalFlowSettingsResource_Schema(t *testing.T) {
	r := NewApprovalFlowSettingsResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	boolAttrs := []string{
		"admin_approval_email",
		"expose_approval_progress",
		"resource_request",
		"user_approval_email",
		"user_request_email",
	}
	for _, name := range boolAttrs {
		attr, ok := attrs[name]
		if !ok {
			t.Errorf("expected attribute %q to exist", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("expected attribute %q to be optional", name)
		}
		if !attr.IsComputed() {
			t.Errorf("expected attribute %q to be computed", name)
		}
	}
}

func TestApprovalFlowSettingsResource_ModelToAPI(t *testing.T) {
	r := &ApprovalFlowSettingsResource{}

	model := &ApprovalFlowSettingsResourceModel{
		AdminApprovalEmail:     types.BoolValue(true),
		ExposeApprovalProgress: types.BoolValue(false),
		ResourceRequest:        types.BoolValue(true),
		UserApprovalEmail:      types.BoolValue(true),
		UserRequestEmail:       types.BoolValue(false),
	}

	settings := r.modelToAPI(model)

	if !settings.AdminApprovalEmail {
		t.Error("expected admin_approval_email true")
	}
	if settings.ExposeApprovalProgress {
		t.Error("expected expose_approval_progress false")
	}
	if !settings.ResourceRequest {
		t.Error("expected resource_request true")
	}
	if !settings.UserApprovalEmail {
		t.Error("expected user_approval_email true")
	}
	if settings.UserRequestEmail {
		t.Error("expected user_request_email false")
	}
}

func TestApprovalFlowSettingsResource_APIToModel(t *testing.T) {
	r := &ApprovalFlowSettingsResource{}

	settings := &client.ApprovalFlowSettings{
		ID:                     "settings-123",
		AdminApprovalEmail:     true,
		ExposeApprovalProgress: true,
		ResourceRequest:        false,
		UserApprovalEmail:      true,
		UserRequestEmail:       true,
	}

	var model ApprovalFlowSettingsResourceModel
	r.apiToModel(settings, &model)

	if model.ID.ValueString() != "settings-123" {
		t.Errorf("expected id 'settings-123', got %q", model.ID.ValueString())
	}
	if !model.AdminApprovalEmail.ValueBool() {
		t.Error("expected admin_approval_email true")
	}
	if !model.ExposeApprovalProgress.ValueBool() {
		t.Error("expected expose_approval_progress true")
	}
	if model.ResourceRequest.ValueBool() {
		t.Error("expected resource_request false")
	}
}
