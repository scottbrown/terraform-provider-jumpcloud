// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

func TestAuthenticationPolicyResource_Metadata(t *testing.T) {
	r := NewAuthenticationPolicyResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_authentication_policy" {
		t.Errorf("expected type name 'jumpcloud_authentication_policy', got %q", resp.TypeName)
	}
}

func TestAuthenticationPolicyResource_Schema(t *testing.T) {
	r := NewAuthenticationPolicyResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"name", "type"}
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

	optional := []string{"disabled", "conditions", "effect", "targets"}
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

func TestAuthenticationPolicyResource_ModelToAPI_Minimal(t *testing.T) {
	r := &AuthenticationPolicyResource{}
	ctx := context.Background()

	model := &AuthenticationPolicyResourceModel{
		Name:       types.StringValue("Block External"),
		Type:       types.StringValue("user_portal"),
		Disabled:   types.BoolValue(false),
		Conditions: types.ObjectNull(conditionsAttrTypes()),
		Effect:     types.ObjectNull(effectAttrTypes()),
		Targets:    types.ListNull(types.ObjectType{AttrTypes: targetAttrTypes()}),
	}

	policy := r.modelToAPI(ctx, model)

	if policy.Name != "Block External" {
		t.Errorf("expected name 'Block External', got %q", policy.Name)
	}
	if policy.Type != "user_portal" {
		t.Errorf("expected type 'user_portal', got %q", policy.Type)
	}
	if policy.Disabled {
		t.Error("expected disabled to be false")
	}
	if policy.Conditions != nil {
		t.Error("expected conditions to be nil")
	}
	if policy.Effect != nil {
		t.Error("expected effect to be nil")
	}
	if policy.Targets != nil {
		t.Error("expected targets to be nil")
	}
}

func TestAuthenticationPolicyResource_APIToModel(t *testing.T) {
	r := &AuthenticationPolicyResource{}

	t.Run("full policy", func(t *testing.T) {
		trueBool := true
		policy := &client.AuthenticationPolicy{
			ID:       "pol-123",
			Name:     "Require MFA from unknown IPs",
			Type:     "user_portal",
			Disabled: false,
			Conditions: &client.AuthenticationPolicyConditions{
				IPAddressIn:   []string{"iplist-abc"},
				DeviceTrusted: &trueBool,
			},
			Effect: &client.AuthenticationPolicyEffect{
				Action: "allow",
				Obligations: &client.AuthenticationPolicyEffectObligations{
					MFA: &client.AuthenticationPolicyMFA{
						Required:       &trueBool,
						AllowedFactors: []string{"totp", "webauthn"},
					},
				},
			},
			Targets: &client.AuthenticationPolicyTargets{
				Resources: []client.AuthenticationPolicyTargetResource{
					{Type: "application", ID: "app-xyz"},
				},
			},
		}

		var model AuthenticationPolicyResourceModel
		r.apiToModel(policy, &model)

		if model.ID.ValueString() != "pol-123" {
			t.Errorf("expected id 'pol-123', got %q", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Require MFA from unknown IPs" {
			t.Errorf("expected name, got %q", model.Name.ValueString())
		}
		if model.Type.ValueString() != "user_portal" {
			t.Errorf("expected type 'user_portal', got %q", model.Type.ValueString())
		}
		if model.Conditions.IsNull() {
			t.Fatal("expected conditions to not be null")
		}
		if model.Effect.IsNull() {
			t.Fatal("expected effect to not be null")
		}
		if model.Targets.IsNull() {
			t.Fatal("expected targets to not be null")
		}
	})

	t.Run("nil nested blocks stay null", func(t *testing.T) {
		policy := &client.AuthenticationPolicy{
			ID:   "pol-456",
			Name: "Simple",
			Type: "resource",
		}

		var model AuthenticationPolicyResourceModel
		r.apiToModel(policy, &model)

		if !model.Conditions.IsNull() {
			t.Error("expected conditions to be null")
		}
		if !model.Effect.IsNull() {
			t.Error("expected effect to be null")
		}
		if !model.Targets.IsNull() {
			t.Error("expected targets to be null")
		}
	})
}
