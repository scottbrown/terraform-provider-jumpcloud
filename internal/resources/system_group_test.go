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

func TestSystemGroupResource_Metadata(t *testing.T) {
	r := NewSystemGroupResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_system_group" {
		t.Errorf("expected type name 'jumpcloud_system_group', got %q", resp.TypeName)
	}
}

func TestSystemGroupResource_Schema(t *testing.T) {
	r := NewSystemGroupResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	if attr, ok := attrs["id"]; !ok {
		t.Error("expected attribute 'id' to exist")
	} else if !attr.IsComputed() {
		t.Error("expected attribute 'id' to be computed")
	}

	if attr, ok := attrs["name"]; !ok {
		t.Error("expected attribute 'name' to exist")
	} else if !attr.IsRequired() {
		t.Error("expected attribute 'name' to be required")
	}

	optional := []string{"description", "membership_method", "member_query"}
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

func TestSystemGroupResource_ModelToAPI(t *testing.T) {
	r := &SystemGroupResource{}

	t.Run("static group", func(t *testing.T) {
		model := SystemGroupResourceModel{
			Name:             types.StringValue("Production"),
			Description:      types.StringValue("Prod servers"),
			MembershipMethod: types.StringValue("STATIC"),
			MemberQuery:      types.StringNull(),
		}

		group := r.modelToAPI(&model)

		if group.Name != "Production" {
			t.Errorf("expected name 'Production', got %q", group.Name)
		}
		if group.Description != "Prod servers" {
			t.Errorf("expected description 'Prod servers', got %q", group.Description)
		}
		if group.MembershipMethod != "STATIC" {
			t.Errorf("expected membership_method 'STATIC', got %q", group.MembershipMethod)
		}
		if group.MemberQuery != nil {
			t.Error("expected member_query to be nil")
		}
	})

	t.Run("dynamic group with filter query", func(t *testing.T) {
		model := SystemGroupResourceModel{
			Name:             types.StringValue("Mac Devices"),
			Description:      types.StringNull(),
			MembershipMethod: types.StringValue("DYNAMIC_AUTOMATED"),
			MemberQuery:      types.StringValue(`{"queryType":"Filter","filters":["osFamily:$eq:darwin"]}`),
		}

		group := r.modelToAPI(&model)

		if group.MembershipMethod != "DYNAMIC_AUTOMATED" {
			t.Errorf("expected membership_method 'DYNAMIC_AUTOMATED', got %q", group.MembershipMethod)
		}
		if group.MemberQuery == nil {
			t.Fatal("expected member_query to not be nil")
		}
		if group.MemberQuery.QueryType != "Filter" {
			t.Errorf("expected query_type 'Filter', got %q", group.MemberQuery.QueryType)
		}
		if len(group.MemberQuery.Filters) != 1 {
			t.Fatalf("expected 1 filter, got %d", len(group.MemberQuery.Filters))
		}
		if group.MemberQuery.Filters[0] != "osFamily:$eq:darwin" {
			t.Errorf("expected filter 'osFamily:$eq:darwin', got %q", group.MemberQuery.Filters[0])
		}
	})

	t.Run("dynamic group with search filters", func(t *testing.T) {
		model := SystemGroupResourceModel{
			Name:             types.StringValue("Tech MacOS"),
			Description:      types.StringNull(),
			MembershipMethod: types.StringValue("DYNAMIC_AUTOMATED"),
			MemberQuery:      types.StringValue(`{"queryType":"Search","searchFilters":"{\"filter\":{\"and\":[\"osFamily:$eq:darwin\"]}}"}`),
		}

		group := r.modelToAPI(&model)

		if group.MemberQuery == nil {
			t.Fatal("expected member_query to not be nil")
		}
		if group.MemberQuery.QueryType != "Search" {
			t.Errorf("expected query_type 'Search', got %q", group.MemberQuery.QueryType)
		}
		if group.MemberQuery.SearchFilters == "" {
			t.Error("expected search_filters to not be empty")
		}
	})
}

func TestSystemGroupResource_APIToModel(t *testing.T) {
	r := &SystemGroupResource{}

	t.Run("static group", func(t *testing.T) {
		group := &client.SystemGroup{
			ID:               "sg-123",
			Name:             "Staging",
			Description:      "Staging servers",
			MembershipMethod: "STATIC",
		}

		var model SystemGroupResourceModel
		r.apiToModel(group, &model)

		if model.ID.ValueString() != "sg-123" {
			t.Errorf("expected id 'sg-123', got %q", model.ID.ValueString())
		}
		if model.Name.ValueString() != "Staging" {
			t.Errorf("expected name 'Staging', got %q", model.Name.ValueString())
		}
		if model.MembershipMethod.ValueString() != "STATIC" {
			t.Errorf("expected membership_method 'STATIC', got %q", model.MembershipMethod.ValueString())
		}
		if !model.MemberQuery.IsNull() {
			t.Error("expected member_query to be null")
		}
	})

	t.Run("dynamic group with filters", func(t *testing.T) {
		group := &client.SystemGroup{
			ID:               "sg-456",
			Name:             "Windows",
			MembershipMethod: "DYNAMIC_AUTOMATED",
			MemberQuery: &client.SystemGroupMemberQuery{
				QueryType: "Filter",
				Filters:   []string{"osFamily:$eq:windows"},
			},
		}

		var model SystemGroupResourceModel
		r.apiToModel(group, &model)

		if model.MemberQuery.IsNull() {
			t.Fatal("expected member_query to not be null")
		}
	})

	t.Run("dynamic group with search filters", func(t *testing.T) {
		group := &client.SystemGroup{
			ID:               "sg-789",
			Name:             "Tech MacOS",
			MembershipMethod: "DYNAMIC_AUTOMATED",
			MemberQuery: &client.SystemGroupMemberQuery{
				QueryType:     "Search",
				SearchFilters: `{"filter":{"and":["osFamily:$eq:darwin","primarySystemUser.department:$in:Tech|Data"]}}`,
			},
		}

		var model SystemGroupResourceModel
		r.apiToModel(group, &model)

		if model.MemberQuery.IsNull() {
			t.Fatal("expected member_query to not be null")
		}
	})

	t.Run("empty optional fields", func(t *testing.T) {
		group := &client.SystemGroup{
			ID:   "sg-789",
			Name: "Minimal",
		}

		var model SystemGroupResourceModel
		r.apiToModel(group, &model)

		if !model.Description.IsNull() {
			t.Errorf("expected description to be null, got %q", model.Description.ValueString())
		}
		if !model.MembershipMethod.IsNull() {
			t.Errorf("expected membership_method to be null, got %q", model.MembershipMethod.ValueString())
		}
	})
}

func TestSystemGroupMembershipResource_Metadata(t *testing.T) {
	r := NewSystemGroupMembershipResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_system_group_membership" {
		t.Errorf("expected type name 'jumpcloud_system_group_membership', got %q", resp.TypeName)
	}
}

func TestSystemGroupMembershipResource_Schema(t *testing.T) {
	r := NewSystemGroupMembershipResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	required := []string{"group_id", "system_id"}
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
