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

func TestCommandAssociationResource_Metadata(t *testing.T) {
	r := NewCommandAssociationResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_command_association" {
		t.Errorf("expected type name 'jumpcloud_command_association', got %q", resp.TypeName)
	}
}

func TestCommandAssociationResource_Schema(t *testing.T) {
	r := NewCommandAssociationResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"command_id", "target_id", "target_type"}
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

func TestCommandAssociationResource_ImportIDParsing(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		wantParts int
		wantErr   bool
	}{
		{"valid", "cmd-123/system_group/sg-456", 3, false},
		{"invalid no separator", "bad-id", 1, true},
		{"invalid one separator", "cmd-123/system", 2, true},
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

func TestCommandAssociationResource_CompositeID(t *testing.T) {
	model := CommandAssociationResourceModel{
		CommandID:  types.StringValue("cmd-abc"),
		TargetType: types.StringValue("system_group"),
		TargetID:   types.StringValue("sg-123"),
	}

	expected := "cmd-abc/system_group/sg-123"
	id := model.CommandID.ValueString() + "/" + model.TargetType.ValueString() + "/" + model.TargetID.ValueString()
	if id != expected {
		t.Errorf("expected composite id %q, got %q", expected, id)
	}
}
