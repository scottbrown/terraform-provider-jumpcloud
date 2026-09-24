// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

func TestRoleResource_Metadata(t *testing.T) {
	r := NewRoleResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_role" {
		t.Errorf("expected type name 'jumpcloud_role', got %q", resp.TypeName)
	}
}

func TestRoleResource_Schema(t *testing.T) {
	r := NewRoleResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"name", "permissions"}
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

	optional := []string{"description", "scope"}
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

func TestRoleResource_APIToModel(t *testing.T) {
	r := &RoleResource{}

	t.Run("full role with scope", func(t *testing.T) {
		role := &client.Role{
			ID:          "role-123",
			Name:        "Helpdesk Admin",
			Description: "Limited admin for helpdesk team",
			Permissions: []client.RolePermission{
				{ResourceType: "user", Action: "read"},
				{ResourceType: "user", Action: "resetPassword"},
			},
			Scope: &client.RoleScope{
				Type: "scope_group",
				ID:   "sg-abc",
			},
		}

		var model RoleResourceModel
		r.apiToModel(role, &model)

		if model.ID.ValueString() != "role-123" {
			t.Errorf("expected id 'role-123', got %q", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Helpdesk Admin" {
			t.Errorf("expected name 'Helpdesk Admin', got %q", model.Name.ValueString())
		}
		if model.Description.ValueString() != "Limited admin for helpdesk team" {
			t.Errorf("expected description, got %q", model.Description.ValueString())
		}
		if model.Permissions.IsNull() {
			t.Fatal("expected permissions to not be null")
		}
		if model.Scope.IsNull() {
			t.Fatal("expected scope to not be null")
		}
	})

	t.Run("minimal role without scope", func(t *testing.T) {
		role := &client.Role{
			ID:   "role-456",
			Name: "Full Admin",
			Permissions: []client.RolePermission{
				{ResourceType: "*", Action: "*"},
			},
		}

		var model RoleResourceModel
		r.apiToModel(role, &model)

		if !model.Description.IsNull() {
			t.Errorf("expected description to be null, got %q", model.Description.ValueString())
		}
		if !model.Scope.IsNull() {
			t.Error("expected scope to be null")
		}
		if model.Permissions.IsNull() {
			t.Fatal("expected permissions to not be null")
		}
	})
}

func TestRoleResource_ModelToAPI(t *testing.T) {
	r := &RoleResource{}
	ctx := context.Background()

	permObjType := types.ObjectType{AttrTypes: rolePermissionAttrTypes()}
	perm1, _ := types.ObjectValue(rolePermissionAttrTypes(), map[string]attr.Value{
		"resource_type": types.StringValue("user"),
		"action":        types.StringValue("read"),
	})
	perms, _ := types.ListValue(permObjType, []attr.Value{perm1})

	scopeObj, _ := types.ObjectValue(roleScopeAttrTypes(), map[string]attr.Value{
		"type": types.StringValue("all"),
		"id":   types.StringNull(),
	})

	model := &RoleResourceModel{
		Name:        types.StringValue("Reader"),
		Description: types.StringValue("Read-only access"),
		Permissions: perms,
		Scope:       scopeObj,
	}

	role := r.modelToAPI(ctx, model)

	if role.Name != "Reader" {
		t.Errorf("expected name 'Reader', got %q", role.Name)
	}
	if role.Description != "Read-only access" {
		t.Errorf("expected description 'Read-only access', got %q", role.Description)
	}
	if len(role.Permissions) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(role.Permissions))
	}
	if role.Permissions[0].ResourceType != "user" {
		t.Errorf("expected resource_type 'user', got %q", role.Permissions[0].ResourceType)
	}
	if role.Permissions[0].Action != "read" {
		t.Errorf("expected action 'read', got %q", role.Permissions[0].Action)
	}
	if role.Scope == nil {
		t.Fatal("expected scope to not be nil")
	}
	if role.Scope.Type != "all" {
		t.Errorf("expected scope type 'all', got %q", role.Scope.Type)
	}
}
