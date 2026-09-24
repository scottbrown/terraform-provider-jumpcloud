// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestUserGroupDataSource_Metadata(t *testing.T) {
	d := NewUserGroupDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_user_group" {
		t.Errorf("expected type name 'jumpcloud_user_group', got %q", resp.TypeName)
	}
}

func TestUserGroupDataSource_Schema(t *testing.T) {
	d := NewUserGroupDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	idAttr, ok := attrs["id"]
	if !ok {
		t.Fatal("expected 'id' attribute")
	}
	if !idAttr.IsOptional() {
		t.Error("expected 'id' to be optional")
	}
	if !idAttr.IsComputed() {
		t.Error("expected 'id' to be computed")
	}

	nameAttr, ok := attrs["name"]
	if !ok {
		t.Fatal("expected 'name' attribute")
	}
	if !nameAttr.IsOptional() {
		t.Error("expected 'name' to be optional")
	}
	if !nameAttr.IsComputed() {
		t.Error("expected 'name' to be computed")
	}

	computed := []string{"description", "membership_method"}
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
