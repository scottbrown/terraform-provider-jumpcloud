// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestCommandDataSource_Metadata(t *testing.T) {
	d := NewCommandDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_command" {
		t.Errorf("expected type name 'jumpcloud_command', got %q", resp.TypeName)
	}
}

func TestCommandDataSource_Schema(t *testing.T) {
	d := NewCommandDataSource()
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

	computed := []string{"name", "command", "command_type", "user", "sudo", "shell", "trigger"}
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
