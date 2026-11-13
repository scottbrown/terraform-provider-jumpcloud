// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestPasswordPolicyDataSource_Metadata(t *testing.T) {
	d := NewPasswordPolicyDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_password_policy" {
		t.Errorf("expected type name 'jumpcloud_password_policy', got %q", resp.TypeName)
	}
}

func TestPasswordPolicyDataSource_Schema(t *testing.T) {
	d := NewPasswordPolicyDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	optionalComputed := []string{"id", "name"}
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

	computed := []string{
		"enable_min_length", "min_length",
		"needs_uppercase", "needs_lowercase", "needs_numeric", "needs_symbolic",
		"enable_max_login_attempts", "max_login_attempts",
		"lockout_time_in_seconds", "password_expiration_in_days",
		"precedence",
	}
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
