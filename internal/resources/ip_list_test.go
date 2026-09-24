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

func TestIPListResource_Metadata(t *testing.T) {
	r := NewIPListResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_ip_list" {
		t.Errorf("expected type name 'jumpcloud_ip_list', got %q", resp.TypeName)
	}
}

func TestIPListResource_Schema(t *testing.T) {
	r := NewIPListResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"name", "ips"}
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

	optional := []string{"description"}
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

func TestIPListResource_ModelToAPI(t *testing.T) {
	r := &IPListResource{}
	ctx := context.Background()

	ips, _ := types.ListValueFrom(ctx, types.StringType, []string{"10.0.0.0/8", "192.168.1.1"})
	model := &IPListResourceModel{
		Name:        types.StringValue("Office IPs"),
		Description: types.StringValue("Corporate office addresses"),
		IPs:         ips,
	}

	ipList := r.modelToAPI(ctx, model)

	if ipList.Name != "Office IPs" {
		t.Errorf("expected name 'Office IPs', got %q", ipList.Name)
	}
	if ipList.Description != "Corporate office addresses" {
		t.Errorf("expected description 'Corporate office addresses', got %q", ipList.Description)
	}
	if len(ipList.IPs) != 2 {
		t.Fatalf("expected 2 IPs, got %d", len(ipList.IPs))
	}
	if ipList.IPs[0] != "10.0.0.0/8" {
		t.Errorf("expected first IP '10.0.0.0/8', got %q", ipList.IPs[0])
	}
	if ipList.IPs[1] != "192.168.1.1" {
		t.Errorf("expected second IP '192.168.1.1', got %q", ipList.IPs[1])
	}
}

func TestIPListResource_APIToModel(t *testing.T) {
	r := &IPListResource{}

	t.Run("full fields", func(t *testing.T) {
		ipList := &client.IPList{
			ID:          "iplist-abc",
			Name:        "VPN Egress",
			Description: "VPN exit nodes",
			IPs:         []string{"203.0.113.0/24", "198.51.100.5"},
		}

		var model IPListResourceModel
		r.apiToModel(ipList, &model)

		if model.ID.ValueString() != "iplist-abc" {
			t.Errorf("expected id 'iplist-abc', got %q", model.ID.ValueString())
		}
		if model.Name.ValueString() != "VPN Egress" {
			t.Errorf("expected name 'VPN Egress', got %q", model.Name.ValueString())
		}
		if model.Description.ValueString() != "VPN exit nodes" {
			t.Errorf("expected description 'VPN exit nodes', got %q", model.Description.ValueString())
		}
		if model.IPs.IsNull() {
			t.Fatal("expected IPs to not be null")
		}
	})

	t.Run("empty description stays null", func(t *testing.T) {
		ipList := &client.IPList{
			ID:   "iplist-xyz",
			Name: "Minimal",
			IPs:  []string{"1.2.3.4"},
		}

		var model IPListResourceModel
		r.apiToModel(ipList, &model)

		if !model.Description.IsNull() {
			t.Errorf("expected description to be null, got %q", model.Description.ValueString())
		}
	})

	t.Run("empty ips returns empty list not null", func(t *testing.T) {
		ipList := &client.IPList{
			ID:   "iplist-empty",
			Name: "Empty",
			IPs:  []string{},
		}

		var model IPListResourceModel
		r.apiToModel(ipList, &model)

		if model.IPs.IsNull() {
			t.Error("expected IPs to be empty list, not null")
		}
	})
}
