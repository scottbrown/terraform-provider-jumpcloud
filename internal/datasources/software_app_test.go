// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestSoftwareAppDataSource_Metadata(t *testing.T) {
	d := NewSoftwareAppDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_software_app" {
		t.Errorf("expected type name 'jumpcloud_software_app', got %q", resp.TypeName)
	}
}

func TestSoftwareAppDataSource_Schema(t *testing.T) {
	d := NewSoftwareAppDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	optionalComputed := []string{"id", "display_name"}
	for _, name := range optionalComputed {
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

	computed := []string{"settings"}
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
