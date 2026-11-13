// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

func TestApprovalFlowResource_Metadata(t *testing.T) {
	r := NewApprovalFlowResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_approval_flow" {
		t.Errorf("expected type name 'jumpcloud_approval_flow', got %q", resp.TypeName)
	}
}

func TestApprovalFlowResource_Schema(t *testing.T) {
	r := NewApprovalFlowResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"name", "approval_type", "resource_id", "resource_type"}
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

	optional := []string{"description", "status", "time_based_access", "non_admin_approval", "visible_to"}
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
}

func TestApprovalFlowResource_ModelToAPI(t *testing.T) {
	r := &ApprovalFlowResource{}

	t.Run("basic fields", func(t *testing.T) {
		model := &ApprovalFlowResourceModel{
			Name:                types.StringValue("Access Request"),
			Description:         types.StringValue("Request access to resources"),
			ApprovalType:        types.StringValue("manual"),
			ApproverRequirement: types.StringValue("at_least_one"),
			Category:            types.StringValue("general"),
			ResourceID:          types.StringValue("grp-123"),
			ResourceType:        types.StringValue("user_group"),
			Status:              types.StringValue("active"),
			TimeBasedAccess:     types.BoolValue(true),
			FixedDuration:       types.BoolValue(false),
			NonAdminApproval:    types.BoolValue(true),
			SlackEnabled:        types.BoolValue(false),
			VisibleTo:           types.ListNull(types.StringType),
		}

		flow := r.modelToAPI(model)

		if flow.Name != "Access Request" {
			t.Errorf("expected name 'Access Request', got %q", flow.Name)
		}
		if flow.ApprovalType != "manual" {
			t.Errorf("expected approval_type 'manual', got %q", flow.ApprovalType)
		}
		if flow.ApproverRequirement != "at_least_one" {
			t.Errorf("expected approver_requirement 'at_least_one', got %q", flow.ApproverRequirement)
		}
		if flow.ResourceID != "grp-123" {
			t.Errorf("expected resource_id 'grp-123', got %q", flow.ResourceID)
		}
		if !flow.TimeBasedAccess {
			t.Error("expected time_based_access to be true")
		}
		if !flow.NonAdminApproval {
			t.Error("expected non_admin_approval to be true")
		}
		if flow.FixedDuration != "" {
			t.Errorf("expected empty fixed_duration, got %q", flow.FixedDuration)
		}
	})

	t.Run("fixed duration enabled", func(t *testing.T) {
		model := &ApprovalFlowResourceModel{
			Name:             types.StringValue("Test"),
			ApprovalType:     types.StringValue("auto"),
			ResourceID:       types.StringValue("grp-1"),
			ResourceType:     types.StringValue("user_group"),
			TimeBasedAccess:  types.BoolValue(false),
			FixedDuration:    types.BoolValue(true),
			NonAdminApproval: types.BoolValue(false),
			SlackEnabled:     types.BoolValue(false),
			VisibleTo:        types.ListNull(types.StringType),
		}

		flow := r.modelToAPI(model)
		if flow.FixedDuration != "P1D" {
			t.Errorf("expected fixed_duration 'P1D' when enabled, got %q", flow.FixedDuration)
		}
	})

	t.Run("visible_to list", func(t *testing.T) {
		elems := []attr.Value{
			types.StringValue("grp-aaa"),
			types.StringValue("grp-bbb"),
		}
		visibleTo, _ := types.ListValue(types.StringType, elems)

		model := &ApprovalFlowResourceModel{
			Name:             types.StringValue("Test"),
			ApprovalType:     types.StringValue("manual"),
			ResourceID:       types.StringValue("grp-1"),
			ResourceType:     types.StringValue("user_group"),
			TimeBasedAccess:  types.BoolValue(false),
			FixedDuration:    types.BoolValue(false),
			NonAdminApproval: types.BoolValue(false),
			SlackEnabled:     types.BoolValue(false),
			VisibleTo:        visibleTo,
		}

		flow := r.modelToAPI(model)
		if len(flow.VisibleTo) != 2 {
			t.Fatalf("expected 2 visible_to entries, got %d", len(flow.VisibleTo))
		}
		if flow.VisibleTo[0].ID != "grp-aaa" {
			t.Errorf("expected first visible_to id 'grp-aaa', got %q", flow.VisibleTo[0].ID)
		}
		if flow.VisibleTo[1].ID != "grp-bbb" {
			t.Errorf("expected second visible_to id 'grp-bbb', got %q", flow.VisibleTo[1].ID)
		}
	})
}

func TestApprovalFlowResource_APIToModel(t *testing.T) {
	r := &ApprovalFlowResource{}

	t.Run("full response", func(t *testing.T) {
		flow := &client.ApprovalFlow{
			ID:                  "flow-1",
			Name:                "Deploy Access",
			Description:         "Access for deploys",
			ApprovalType:        "manual",
			ApproverRequirement: "all",
			Category:            "engineering",
			Status:              "active",
			TimeBasedAccess:     true,
			FixedDuration:       "P1D",
			NonAdminApproval:    false,
			SlackEnabled:        true,
			Resources: []client.ApprovalFlowResource{
				{ResourceID: "grp-deploy", ResourceType: "user_group"},
			},
			VisibleTo: []client.VisibleToEntry{
				{ID: "grp-vis-1"},
				{ID: "grp-vis-2"},
			},
		}

		var model ApprovalFlowResourceModel
		r.apiToModel(flow, &model)

		if model.Name.ValueString() != "Deploy Access" {
			t.Errorf("expected name 'Deploy Access', got %q", model.Name.ValueString())
		}
		if model.Description.ValueString() != "Access for deploys" {
			t.Errorf("expected description 'Access for deploys', got %q", model.Description.ValueString())
		}
		if model.ResourceID.ValueString() != "grp-deploy" {
			t.Errorf("expected resource_id 'grp-deploy', got %q", model.ResourceID.ValueString())
		}
		if model.ResourceType.ValueString() != "user_group" {
			t.Errorf("expected resource_type 'user_group', got %q", model.ResourceType.ValueString())
		}
		if !model.TimeBasedAccess.ValueBool() {
			t.Error("expected time_based_access true")
		}
		if !model.FixedDuration.ValueBool() {
			t.Error("expected fixed_duration true when API returns non-empty string")
		}
		if !model.SlackEnabled.ValueBool() {
			t.Error("expected slack_enabled true")
		}

		var visibleIDs []string
		model.VisibleTo.ElementsAs(context.Background(), &visibleIDs, false)
		if len(visibleIDs) != 2 {
			t.Fatalf("expected 2 visible_to ids, got %d", len(visibleIDs))
		}
		if visibleIDs[0] != "grp-vis-1" || visibleIDs[1] != "grp-vis-2" {
			t.Errorf("unexpected visible_to ids: %v", visibleIDs)
		}
	})

	t.Run("empty description becomes null", func(t *testing.T) {
		flow := &client.ApprovalFlow{
			ID:           "flow-2",
			Name:         "Minimal",
			ApprovalType: "auto",
			Status:       "active",
			Resources: []client.ApprovalFlowResource{
				{ResourceID: "grp-1", ResourceType: "user_group"},
			},
		}

		var model ApprovalFlowResourceModel
		r.apiToModel(flow, &model)

		if !model.Description.IsNull() {
			t.Errorf("expected description to be null, got %q", model.Description.ValueString())
		}
	})

	t.Run("empty visible_to becomes null list", func(t *testing.T) {
		flow := &client.ApprovalFlow{
			ID:           "flow-3",
			Name:         "No Visibility",
			ApprovalType: "manual",
			Status:       "active",
			Resources: []client.ApprovalFlowResource{
				{ResourceID: "grp-1", ResourceType: "user_group"},
			},
		}

		var model ApprovalFlowResourceModel
		r.apiToModel(flow, &model)

		if !model.VisibleTo.IsNull() {
			t.Error("expected visible_to to be null list when empty")
		}
	})
}
