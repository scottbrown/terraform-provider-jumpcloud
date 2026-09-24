// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"jumpcloud": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck verifies the environment is ready for acceptance tests.
// Acceptance tests create real resources in JumpCloud and require an API key.
// Tests are skipped rather than failed when credentials are absent so that
// CI runs without secrets (for example, pull requests from forks) stay green.
func testAccPreCheck(t *testing.T) {
	if os.Getenv("JUMPCLOUD_API_KEY") == "" {
		t.Skip("JUMPCLOUD_API_KEY must be set for acceptance tests")
	}
}

func TestProvider_Metadata(t *testing.T) {
	p := New("1.0.0")()
	req := provider.MetadataRequest{}
	var resp provider.MetadataResponse
	p.Metadata(t.Context(), req, &resp)

	if resp.TypeName != "jumpcloud" {
		t.Errorf("expected type name 'jumpcloud', got %q", resp.TypeName)
	}
	if resp.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", resp.Version)
	}
}

func TestProvider_Schema(t *testing.T) {
	p := New("0.1.0")()
	req := provider.SchemaRequest{}
	var resp provider.SchemaResponse
	p.Schema(t.Context(), req, &resp)

	attrs := resp.Schema.Attributes
	expected := []string{"api_key", "org_id", "base_url"}
	for _, name := range expected {
		attr, ok := attrs[name]
		if !ok {
			t.Errorf("expected attribute %q to exist", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("expected attribute %q to be optional", name)
		}
	}

	if !attrs["api_key"].IsSensitive() {
		t.Error("expected attribute \"api_key\" to be sensitive")
	}
}

func TestProvider_Resources(t *testing.T) {
	p, ok := New("0.1.0")().(*JumpCloudProvider)
	if !ok {
		t.Fatal("expected *JumpCloudProvider")
	}
	res := p.Resources(t.Context())

	expectedCount := 23
	if len(res) != expectedCount {
		t.Errorf("expected %d resources, got %d", expectedCount, len(res))
	}
}

func TestProvider_DataSources(t *testing.T) {
	p, ok := New("0.1.0")().(*JumpCloudProvider)
	if !ok {
		t.Fatal("expected *JumpCloudProvider")
	}
	ds := p.DataSources(t.Context())

	expectedCount := 15
	if len(ds) != expectedCount {
		t.Errorf("expected %d data sources, got %d", expectedCount, len(ds))
	}
}
