// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

func TestPasswordPolicyResource_Metadata(t *testing.T) {
	r := NewPasswordPolicyResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_password_policy" {
		t.Errorf("expected type name 'jumpcloud_password_policy', got %q", resp.TypeName)
	}
}

func TestPasswordPolicyResource_Schema(t *testing.T) {
	r := NewPasswordPolicyResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	computed := []string{"id"}
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

	optional := []string{
		"name", "min_length", "enable_min_length",
		"needs_uppercase", "needs_lowercase", "needs_numeric", "needs_symbolic",
		"enable_max_login_attempts", "max_login_attempts",
		"enable_lockout_time_in_seconds", "lockout_time_in_seconds",
		"enable_password_expiration_in_days", "password_expiration_in_days",
		"enable_max_history", "max_history",
		"precedence",
	}
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

func TestPasswordPolicyResource_ModelToAPI(t *testing.T) {
	r := &PasswordPolicyResource{}

	model := PasswordPolicyResourceModel{
		Name:                     types.StringValue("Strong Policy"),
		EnableMinLength:          types.BoolValue(true),
		MinLength:                types.Int64Value(12),
		NeedsUppercase:           types.BoolValue(true),
		NeedsLowercase:           types.BoolValue(true),
		NeedsNumeric:             types.BoolValue(true),
		NeedsSymbolic:            types.BoolValue(false),
		EnableMaxLoginAttempts:   types.BoolValue(true),
		MaxLoginAttempts:         types.Int64Value(5),
		LockoutTimeInSeconds:     types.Int64Value(1800),
		PasswordExpirationInDays: types.Int64Value(90),
		EnableMaxHistory:         types.BoolValue(true),
		MaxHistory:               types.Int64Value(5),
		Precedence:               types.Int64Value(1),
	}

	policy := r.modelToAPI(&model)

	if policy.Name != "Strong Policy" {
		t.Errorf("expected name 'Strong Policy', got %q", policy.Name)
	}
	if policy.MinLength != 12 {
		t.Errorf("expected min_length 12, got %d", policy.MinLength)
	}
	if !policy.EnableMinLength {
		t.Error("expected enable_min_length to be true")
	}
	if !policy.NeedsUppercase {
		t.Error("expected needs_uppercase to be true")
	}
	if policy.NeedsSymbolic {
		t.Error("expected needs_symbolic to be false")
	}
	if policy.MaxLoginAttempts != 5 {
		t.Errorf("expected max_login_attempts 5, got %d", policy.MaxLoginAttempts)
	}
	if policy.LockoutTimeInSeconds != 1800 {
		t.Errorf("expected lockout_time_in_seconds 1800, got %d", policy.LockoutTimeInSeconds)
	}
	if policy.Precedence != 1 {
		t.Errorf("expected precedence 1, got %d", policy.Precedence)
	}
}

func TestPasswordPolicyResource_APIToModel(t *testing.T) {
	r := &PasswordPolicyResource{}

	t.Run("full fields", func(t *testing.T) {
		policy := &client.PasswordPolicy{
			ID:                         "pol-123",
			Name:                       "Strong Policy",
			EnableMinLength:            true,
			MinLength:                  12,
			NeedsUppercase:             true,
			NeedsLowercase:             true,
			NeedsNumeric:               true,
			NeedsSymbolic:              false,
			EnableMaxLoginAttempts:     true,
			MaxLoginAttempts:           5,
			EnableLockoutTimeInSeconds: true,
			LockoutTimeInSeconds:       1800,
			EnableMaxHistory:           true,
			MaxHistory:                 5,
			Precedence:                 1,
		}

		var model PasswordPolicyResourceModel
		r.apiToModel(policy, &model)

		if model.Name.ValueString() != "Strong Policy" {
			t.Errorf("expected name 'Strong Policy', got %q", model.Name.ValueString())
		}
		if model.MinLength.ValueInt64() != 12 {
			t.Errorf("expected min_length 12, got %d", model.MinLength.ValueInt64())
		}
		if !model.NeedsUppercase.ValueBool() {
			t.Error("expected needs_uppercase to be true")
		}
		if model.NeedsSymbolic.ValueBool() {
			t.Error("expected needs_symbolic to be false")
		}
		if model.Precedence.ValueInt64() != 1 {
			t.Errorf("expected precedence 1, got %d", model.Precedence.ValueInt64())
		}
	})

	t.Run("zero optional int fields are null", func(t *testing.T) {
		policy := &client.PasswordPolicy{
			ID:   "pol-456",
			Name: "Minimal Policy",
		}

		var model PasswordPolicyResourceModel
		r.apiToModel(policy, &model)

		if !model.MinLength.IsNull() {
			t.Errorf("expected min_length to be null, got %d", model.MinLength.ValueInt64())
		}
		if !model.MaxLoginAttempts.IsNull() {
			t.Errorf("expected max_login_attempts to be null, got %d", model.MaxLoginAttempts.ValueInt64())
		}
		if !model.Precedence.IsNull() {
			t.Errorf("expected precedence to be null, got %d", model.Precedence.ValueInt64())
		}
	})

	t.Run("booleans are always set", func(t *testing.T) {
		policy := &client.PasswordPolicy{
			ID:   "pol-789",
			Name: "Bool Test",
		}

		var model PasswordPolicyResourceModel
		r.apiToModel(policy, &model)

		if model.NeedsUppercase.IsNull() {
			t.Error("expected needs_uppercase to not be null")
		}
		if model.NeedsUppercase.ValueBool() {
			t.Error("expected needs_uppercase to be false")
		}
		if model.EnableMinLength.IsNull() {
			t.Error("expected enable_min_length to not be null")
		}
	})
}
