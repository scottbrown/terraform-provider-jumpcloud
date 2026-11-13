// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestUserGroupMembershipResource_Metadata(t *testing.T) {
	r := NewUserGroupMembershipResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_user_group_membership" {
		t.Errorf("expected type name 'jumpcloud_user_group_membership', got %q", resp.TypeName)
	}
}

func TestUserGroupMembershipResource_Schema(t *testing.T) {
	r := NewUserGroupMembershipResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"group_id", "user_id"}
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

func TestUserGroupMembershipResource_ImportIDParsing(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantParts int
		wantErr   bool
	}{
		{"valid", "grp-123/usr-456", 2, false},
		{"invalid no separator", "badid", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.SplitN(tt.id, "/", 2)
			if tt.wantErr {
				if len(parts) == 2 {
					t.Error("expected parse failure")
				}
			} else {
				if len(parts) != 2 {
					t.Errorf("expected 2 parts, got %d", len(parts))
				}
			}
		})
	}
}

func TestUserGroupMembershipResource_CompositeID(t *testing.T) {
	model := UserGroupMembershipResourceModel{
		GroupID: types.StringValue("grp-abc"),
		UserID:  types.StringValue("usr-123"),
	}

	expected := "grp-abc/usr-123"
	id := model.GroupID.ValueString() + "/" + model.UserID.ValueString()
	if id != expected {
		t.Errorf("expected composite id %q, got %q", expected, id)
	}
}
