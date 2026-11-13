// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

func TestPolicyResource_Metadata(t *testing.T) {
	r := NewPolicyResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_policy" {
		t.Errorf("expected type name 'jumpcloud_policy', got %q", resp.TypeName)
	}
}

func TestPolicyResource_Schema(t *testing.T) {
	r := NewPolicyResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	if _, ok := attrs["name"]; !ok {
		t.Error("expected 'name' attribute")
	}
	if !attrs["name"].IsRequired() {
		t.Error("expected 'name' to be required")
	}

	if _, ok := attrs["template_id"]; !ok {
		t.Error("expected 'template_id' attribute")
	}
	if !attrs["template_id"].IsOptional() {
		t.Error("expected 'template_id' to be optional")
	}

	if !attrs["values"].IsOptional() {
		t.Error("expected 'values' to be optional")
	}
	if !attrs["values"].IsComputed() {
		t.Error("expected 'values' to be computed")
	}
}

func TestPolicyResource_ModelToAPI(t *testing.T) {
	r := &PolicyResource{}

	t.Run("all fields set", func(t *testing.T) {
		model := &PolicyResourceModel{
			Name:       types.StringValue("MDM Policy"),
			TemplateID: types.StringValue("tmpl-123"),
			State:      types.StringValue("active"),
			Notes:      types.StringValue("Test notes"),
			Values:     types.StringValue(`[{"configFieldName":"payload","value":"data"}]`),
		}

		policy := r.modelToAPI(model)
		if policy.Name != "MDM Policy" {
			t.Errorf("expected name 'MDM Policy', got %q", policy.Name)
		}
		if policy.TemplateID != "tmpl-123" {
			t.Errorf("expected template_id 'tmpl-123', got %q", policy.TemplateID)
		}
		if policy.State != "active" {
			t.Errorf("expected state 'active', got %q", policy.State)
		}
		if policy.Notes != "Test notes" {
			t.Errorf("expected notes 'Test notes', got %q", policy.Notes)
		}
		if string(policy.Values) != `[{"configFieldName":"payload","value":"data"}]` {
			t.Errorf("unexpected values: %s", string(policy.Values))
		}
	})

	t.Run("null optional fields omitted", func(t *testing.T) {
		model := &PolicyResourceModel{
			Name:       types.StringValue("Minimal"),
			TemplateID: types.StringNull(),
			State:      types.StringNull(),
			Notes:      types.StringNull(),
			Values:     types.StringNull(),
		}

		policy := r.modelToAPI(model)
		if policy.TemplateID != "" {
			t.Errorf("expected empty template_id, got %q", policy.TemplateID)
		}
		if policy.State != "" {
			t.Errorf("expected empty state, got %q", policy.State)
		}
		if policy.Values != nil {
			t.Errorf("expected nil values, got %s", string(policy.Values))
		}
	})
}

func TestPolicyResource_APIToModel(t *testing.T) {
	r := &PolicyResource{}

	t.Run("normalizes JSON values", func(t *testing.T) {
		policy := &client.Policy{
			ID:         "pol-1",
			Name:       "Custom MDM",
			TemplateID: "tmpl-abc",
			State:      "active",
			Notes:      "Some notes",
			Values:     json.RawMessage(`[{"configFieldName":"payload","value":"test"}]`),
		}

		var model PolicyResourceModel
		r.apiToModel(policy, &model)

		if model.Name.ValueString() != "Custom MDM" {
			t.Errorf("expected name 'Custom MDM', got %q", model.Name.ValueString())
		}
		if model.TemplateID.ValueString() != "tmpl-abc" {
			t.Errorf("expected template_id 'tmpl-abc', got %q", model.TemplateID.ValueString())
		}
		if model.State.ValueString() != "active" {
			t.Errorf("expected state 'active', got %q", model.State.ValueString())
		}

		var parsed any
		if err := json.Unmarshal([]byte(model.Values.ValueString()), &parsed); err != nil {
			t.Fatalf("values should be valid JSON: %v", err)
		}
	})

	t.Run("empty state becomes null", func(t *testing.T) {
		policy := &client.Policy{
			ID:   "pol-2",
			Name: "No State",
		}

		var model PolicyResourceModel
		r.apiToModel(policy, &model)

		if !model.State.IsNull() {
			t.Errorf("expected state to be null, got %q", model.State.ValueString())
		}
	})

	t.Run("empty values becomes null", func(t *testing.T) {
		policy := &client.Policy{
			ID:   "pol-3",
			Name: "No Values",
		}

		var model PolicyResourceModel
		r.apiToModel(policy, &model)

		if !model.Values.IsNull() {
			t.Errorf("expected values to be null, got %q", model.Values.ValueString())
		}
	})

	t.Run("null JSON values becomes null", func(t *testing.T) {
		policy := &client.Policy{
			ID:     "pol-4",
			Name:   "Null Values",
			Values: json.RawMessage(`null`),
		}

		var model PolicyResourceModel
		r.apiToModel(policy, &model)

		if !model.Values.IsNull() {
			t.Errorf("expected values to be null for JSON null, got %q", model.Values.ValueString())
		}
	})

	t.Run("empty template_id stays empty", func(t *testing.T) {
		policy := &client.Policy{
			ID:   "pol-5",
			Name: "No Template",
		}

		var model PolicyResourceModel
		r.apiToModel(policy, &model)

		if model.TemplateID.IsNull() {
			// expected: empty string means the field is not populated
		} else if model.TemplateID.ValueString() != "" {
			t.Errorf("expected empty or null template_id, got %q", model.TemplateID.ValueString())
		}
	})

	t.Run("notes null preserved", func(t *testing.T) {
		policy := &client.Policy{
			ID:   "pol-6",
			Name: "No Notes",
		}

		model := PolicyResourceModel{
			Notes: types.StringNull(),
		}
		r.apiToModel(policy, &model)

		if !model.Notes.IsNull() {
			t.Errorf("expected notes to stay null, got %q", model.Notes.ValueString())
		}
	})
}
