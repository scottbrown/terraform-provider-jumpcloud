// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestPolicyTemplateDataSource_Metadata(t *testing.T) {
	d := NewPolicyTemplateDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_policy_template" {
		t.Errorf("expected type name 'jumpcloud_policy_template', got %q", resp.TypeName)
	}
}

func TestPolicyTemplateDataSource_Schema(t *testing.T) {
	d := NewPolicyTemplateDataSource()
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

	computed := []string{"name", "description", "os_family"}
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
