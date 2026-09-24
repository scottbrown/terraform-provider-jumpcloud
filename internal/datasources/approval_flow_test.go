// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestApprovalFlowDataSource_Metadata(t *testing.T) {
	d := NewApprovalFlowDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_approval_flow" {
		t.Errorf("expected type name 'jumpcloud_approval_flow', got %q", resp.TypeName)
	}
}

func TestApprovalFlowDataSource_Schema(t *testing.T) {
	d := NewApprovalFlowDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	idAttr, ok := attrs["id"]
	if !ok {
		t.Fatal("expected 'id' attribute")
	}
	if !idAttr.IsRequired() {
		t.Error("expected 'id' to be required")
	}

	computed := []string{"name", "description", "approval_type", "resource_id", "resource_type", "status", "time_based_access", "non_admin_approval"}
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
}
