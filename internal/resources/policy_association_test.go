// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPolicyAssociationResource_Metadata(t *testing.T) {
	r := NewPolicyAssociationResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_policy_association" {
		t.Errorf("expected type name 'jumpcloud_policy_association', got %q", resp.TypeName)
	}
}

func TestPolicyAssociationResource_Schema(t *testing.T) {
	r := NewPolicyAssociationResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"policy_id", "target_id", "target_type"}
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

func TestPolicyAssociationResource_ImportIDParsing(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantParts int
		wantErr   bool
	}{
		{"valid", "pol-123/system_group/sg-456", 3, false},
		{"invalid no separator", "bad-id", 1, true},
		{"invalid one separator", "pol-123/system", 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := strings.SplitN(tt.id, "/", 3)
			if tt.wantErr {
				if len(parts) == 3 {
					t.Error("expected parse failure")
				}
			} else {
				if len(parts) != 3 {
					t.Errorf("expected 3 parts, got %d", len(parts))
				}
			}
		})
	}
}

func TestPolicyAssociationResource_CompositeID(t *testing.T) {
	model := PolicyAssociationResourceModel{
		PolicyID:   types.StringValue("pol-abc"),
		TargetType: types.StringValue("system"),
		TargetID:   types.StringValue("sys-123"),
	}

	expected := "pol-abc/system/sys-123"
	id := model.PolicyID.ValueString() + "/" + model.TargetType.ValueString() + "/" + model.TargetID.ValueString()
	if id != expected {
		t.Errorf("expected composite id %q, got %q", expected, id)
	}
}
