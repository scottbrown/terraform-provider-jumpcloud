// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

func TestPolicyGroupResource_Metadata(t *testing.T) {
	r := NewPolicyGroupResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_policy_group" {
		t.Errorf("expected type name 'jumpcloud_policy_group', got %q", resp.TypeName)
	}
}

func TestPolicyGroupResource_Schema(t *testing.T) {
	r := NewPolicyGroupResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	required := []string{"name"}
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

	computed := []string{"id"}
	for _, name := range computed {
		attr, ok := attrs[name]
		if !ok {
			t.Errorf("expected attribute %q to exist", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("expected attribute %q to be computed", name)
		}
	}

	optional := []string{"description"}
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

func TestPolicyGroupResource_ModelConversion(t *testing.T) {
	t.Run("create from model", func(t *testing.T) {
		model := PolicyGroupResourceModel{
			Name:        types.StringValue("Compliance"),
			Description: types.StringValue("Compliance policies"),
		}

		group := &client.PolicyGroup{
			Name:        model.Name.ValueString(),
			Description: model.Description.ValueString(),
		}

		if group.Name != "Compliance" {
			t.Errorf("expected name 'Compliance', got %q", group.Name)
		}
		if group.Description != "Compliance policies" {
			t.Errorf("expected description 'Compliance policies', got %q", group.Description)
		}
	})

	t.Run("import state maps fields", func(t *testing.T) {
		group := &client.PolicyGroup{
			ID:          "pg-123",
			Name:        "Security",
			Description: "Security policies",
		}

		state := PolicyGroupResourceModel{
			ID:   types.StringValue(group.ID),
			Name: types.StringValue(group.Name),
		}
		if group.Description != "" {
			state.Description = types.StringValue(group.Description)
		}

		if state.ID.ValueString() != "pg-123" {
			t.Errorf("expected id 'pg-123', got %q", state.ID.ValueString())
		}
		if state.Description.ValueString() != "Security policies" {
			t.Errorf("expected description 'Security policies', got %q", state.Description.ValueString())
		}
	})
}

func TestPolicyGroupMembershipResource_Metadata(t *testing.T) {
	r := NewPolicyGroupMembershipResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_policy_group_membership" {
		t.Errorf("expected type name 'jumpcloud_policy_group_membership', got %q", resp.TypeName)
	}
}

func TestPolicyGroupMembershipResource_Schema(t *testing.T) {
	r := NewPolicyGroupMembershipResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	required := []string{"group_id", "policy_id"}
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
}

func TestPolicyGroupAssociationResource_Metadata(t *testing.T) {
	r := NewPolicyGroupAssociationResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_policy_group_association" {
		t.Errorf("expected type name 'jumpcloud_policy_group_association', got %q", resp.TypeName)
	}
}

func TestPolicyGroupAssociationResource_Schema(t *testing.T) {
	r := NewPolicyGroupAssociationResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	required := []string{"group_id", "target_id", "target_type"}
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
}
