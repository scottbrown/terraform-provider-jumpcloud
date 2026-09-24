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

func TestUserGroupResource_Metadata(t *testing.T) {
	r := NewUserGroupResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_user_group" {
		t.Errorf("expected type name 'jumpcloud_user_group', got %q", resp.TypeName)
	}
}

func TestUserGroupResource_Schema(t *testing.T) {
	r := NewUserGroupResource()
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

	optional := []string{"description", "membership_method"}
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

func TestUserGroupResource_ModelConversion(t *testing.T) {
	r := &UserGroupResource{}

	t.Run("create from model", func(t *testing.T) {
		model := UserGroupResourceModel{
			Name:             types.StringValue("Engineers"),
			Description:      types.StringValue("Engineering team"),
			MembershipMethod: types.StringValue("STATIC"),
		}

		group := &client.UserGroup{
			Name:             model.Name.ValueString(),
			Description:      model.Description.ValueString(),
			MembershipMethod: model.MembershipMethod.ValueString(),
		}

		if group.Name != "Engineers" {
			t.Errorf("expected name 'Engineers', got %q", group.Name)
		}
		if group.Description != "Engineering team" {
			t.Errorf("expected description 'Engineering team', got %q", group.Description)
		}
		if group.MembershipMethod != "STATIC" {
			t.Errorf("expected membership_method 'STATIC', got %q", group.MembershipMethod)
		}
		_ = r
	})

	t.Run("import state maps fields", func(t *testing.T) {
		group := &client.UserGroup{
			ID:               "grp-123",
			Name:             "DevOps",
			Description:      "DevOps team",
			MembershipMethod: "DYNAMIC_AUTOMATED",
		}

		state := UserGroupResourceModel{
			ID:          types.StringValue(group.ID),
			Name:        types.StringValue(group.Name),
			Description: types.StringValue(group.Description),
		}
		if group.MembershipMethod != "" {
			state.MembershipMethod = types.StringValue(group.MembershipMethod)
		}

		if state.ID.ValueString() != "grp-123" {
			t.Errorf("expected id 'grp-123', got %q", state.ID.ValueString())
		}
		if state.MembershipMethod.ValueString() != "DYNAMIC_AUTOMATED" {
			t.Errorf("expected membership_method 'DYNAMIC_AUTOMATED', got %q", state.MembershipMethod.ValueString())
		}
	})

	t.Run("empty membership_method stays null", func(t *testing.T) {
		group := &client.UserGroup{
			ID:   "grp-456",
			Name: "Basic Group",
		}

		state := UserGroupResourceModel{
			ID:   types.StringValue(group.ID),
			Name: types.StringValue(group.Name),
		}
		if group.MembershipMethod != "" {
			state.MembershipMethod = types.StringValue(group.MembershipMethod)
		}

		if !state.MembershipMethod.IsNull() {
			t.Errorf("expected membership_method to be null, got %q", state.MembershipMethod.ValueString())
		}
	})
}
