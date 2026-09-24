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

func TestApplicationResource_Metadata(t *testing.T) {
	r := NewApplicationResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_application" {
		t.Errorf("expected type name 'jumpcloud_application', got %q", resp.TypeName)
	}
}

func TestApplicationResource_Schema(t *testing.T) {
	r := NewApplicationResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	if _, ok := attrs["name"]; !ok {
		t.Error("expected attribute 'name' to exist")
	}
	if !attrs["name"].IsRequired() {
		t.Error("expected attribute 'name' to be required")
	}

	optional := []string{"display_label", "sso_url", "logo", "sso", "provision"}
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

func TestApplicationResource_ModelToAPI_Minimal(t *testing.T) {
	r := &ApplicationResource{}
	ctx := context.Background()

	model := &ApplicationResourceModel{
		Name:         types.StringValue("My App"),
		DisplayLabel: types.StringNull(),
		SsoURL:       types.StringNull(),
		Logo:         types.ObjectNull(appLogoAttrTypes()),
		SSO:          types.ObjectNull(appSSOAttrTypes()),
		Provision:    types.ObjectNull(appProvisionAttrTypes()),
	}

	app := r.modelToAPI(ctx, model)

	if app.Name != "My App" {
		t.Errorf("expected name 'My App', got %q", app.Name)
	}
	if app.DisplayLabel != "" {
		t.Errorf("expected empty display_label, got %q", app.DisplayLabel)
	}
	if app.Logo != nil {
		t.Error("expected logo to be nil")
	}
	if app.SSO != nil {
		t.Error("expected sso to be nil")
	}
	if app.Provision != nil {
		t.Error("expected provision to be nil")
	}
}

func TestApplicationResource_APIToModel(t *testing.T) {
	r := &ApplicationResource{}

	t.Run("full application", func(t *testing.T) {
		app := &client.Application{
			ID:           "app-123",
			Name:         "Slack",
			DisplayLabel: "Slack SSO",
			SsoURL:       "https://slack.com/sso",
			Logo:         &client.ApplicationLogo{URL: "https://img.example.com/slack.png"},
			SSO: &client.ApplicationSSO{
				Type:        "saml",
				SpEntityID:  "https://slack.com/entity",
				AcsURL:      "https://slack.com/acs",
				IdpEntityID: "https://jumpcloud.com/entity",
			},
		}

		var model ApplicationResourceModel
		r.apiToModel(app, &model)

		if model.ID.ValueString() != "app-123" {
			t.Errorf("expected id 'app-123', got %q", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Slack" {
			t.Errorf("expected name 'Slack', got %q", model.Name.ValueString())
		}
		if model.DisplayLabel.ValueString() != "Slack SSO" {
			t.Errorf("expected display_label 'Slack SSO', got %q", model.DisplayLabel.ValueString())
		}
		if model.Logo.IsNull() {
			t.Error("expected logo to not be null")
		}
		if model.SSO.IsNull() {
			t.Error("expected sso to not be null")
		}
		if !model.Provision.IsNull() {
			t.Error("expected provision to be null")
		}
	})

	t.Run("minimal application", func(t *testing.T) {
		app := &client.Application{
			ID:   "app-456",
			Name: "Simple App",
		}

		var model ApplicationResourceModel
		r.apiToModel(app, &model)

		if !model.DisplayLabel.IsNull() {
			t.Error("expected display_label to be null")
		}
		if !model.SsoURL.IsNull() {
			t.Error("expected sso_url to be null")
		}
		if !model.Logo.IsNull() {
			t.Error("expected logo to be null")
		}
		if !model.SSO.IsNull() {
			t.Error("expected sso to be null")
		}
		if !model.Provision.IsNull() {
			t.Error("expected provision to be null")
		}
	})
}

func TestApplicationAssociationResource_Metadata(t *testing.T) {
	r := NewApplicationAssociationResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_application_association" {
		t.Errorf("expected type name 'jumpcloud_application_association', got %q", resp.TypeName)
	}
}

func TestApplicationAssociationResource_Schema(t *testing.T) {
	r := NewApplicationAssociationResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"application_id", "target_id", "target_type"}
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
