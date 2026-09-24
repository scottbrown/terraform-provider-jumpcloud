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

func TestSoftwareAppResource_Metadata(t *testing.T) {
	r := NewSoftwareAppResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_software_app" {
		t.Errorf("expected type name 'jumpcloud_software_app', got %q", resp.TypeName)
	}
}

func TestSoftwareAppResource_Schema(t *testing.T) {
	r := NewSoftwareAppResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	if attr, ok := attrs["id"]; !ok {
		t.Error("expected attribute 'id' to exist")
	} else if !attr.IsComputed() {
		t.Error("expected attribute 'id' to be computed")
	}

	if attr, ok := attrs["display_name"]; !ok {
		t.Error("expected attribute 'display_name' to exist")
	} else if !attr.IsRequired() {
		t.Error("expected attribute 'display_name' to be required")
	}

	if attr, ok := attrs["settings"]; !ok {
		t.Error("expected attribute 'settings' to exist")
	} else {
		if !attr.IsOptional() {
			t.Error("expected attribute 'settings' to be optional")
		}
		if !attr.IsComputed() {
			t.Error("expected attribute 'settings' to be computed")
		}
	}
}

func TestSoftwareAppResource_ModelToAPI(t *testing.T) {
	r := &SoftwareAppResource{}

	model := SoftwareAppResourceModel{
		DisplayName: types.StringValue("Google Chrome"),
		Settings:    types.StringValue(`{"packageId":"com.google.chrome"}`),
	}

	app := r.modelToAPI(&model)

	if app.DisplayName != "Google Chrome" {
		t.Errorf("expected display_name 'Google Chrome', got %q", app.DisplayName)
	}
	if string(app.Settings) != `{"packageId":"com.google.chrome"}` {
		t.Errorf("expected settings to match, got %q", string(app.Settings))
	}
}

func TestSoftwareAppResource_ModelToAPI_NullSettings(t *testing.T) {
	r := &SoftwareAppResource{}

	model := SoftwareAppResourceModel{
		DisplayName: types.StringValue("Slack"),
		Settings:    types.StringNull(),
	}

	app := r.modelToAPI(&model)

	if app.DisplayName != "Slack" {
		t.Errorf("expected display_name 'Slack', got %q", app.DisplayName)
	}
	if app.Settings != nil {
		t.Errorf("expected nil settings, got %q", string(app.Settings))
	}
}

func TestSoftwareAppResource_APIToModel(t *testing.T) {
	r := &SoftwareAppResource{}

	t.Run("with settings", func(t *testing.T) {
		app := &client.SoftwareApp{
			ID:          "app-123",
			DisplayName: "Firefox",
			Settings:    []byte(`{"packageId": "org.mozilla.firefox"}`),
		}

		var model SoftwareAppResourceModel
		r.apiToModel(app, &model)

		if model.ID.ValueString() != "app-123" {
			t.Errorf("expected id 'app-123', got %q", model.ID.ValueString())
		}
		if model.DisplayName.ValueString() != "Firefox" {
			t.Errorf("expected display_name 'Firefox', got %q", model.DisplayName.ValueString())
		}
		if model.Settings.ValueString() != `{"packageId":"org.mozilla.firefox"}` {
			t.Errorf("expected compact JSON settings, got %q", model.Settings.ValueString())
		}
	})

	t.Run("null settings", func(t *testing.T) {
		app := &client.SoftwareApp{
			ID:          "app-456",
			DisplayName: "Zoom",
		}

		var model SoftwareAppResourceModel
		r.apiToModel(app, &model)

		if !model.Settings.IsNull() {
			t.Errorf("expected settings to be null, got %q", model.Settings.ValueString())
		}
	})
}

func TestSoftwareAppAssociationResource_Metadata(t *testing.T) {
	r := NewSoftwareAppAssociationResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_software_app_association" {
		t.Errorf("expected type name 'jumpcloud_software_app_association', got %q", resp.TypeName)
	}
}

func TestSoftwareAppAssociationResource_Schema(t *testing.T) {
	r := NewSoftwareAppAssociationResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	required := []string{"software_app_id", "target_id", "target_type"}
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

	if attr, ok := attrs["id"]; !ok {
		t.Error("expected attribute 'id' to exist")
	} else if !attr.IsComputed() {
		t.Error("expected attribute 'id' to be computed")
	}
}
