// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestScheduledReportDataSource_Metadata(t *testing.T) {
	d := NewScheduledReportDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp datasource.MetadataResponse
	d.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_scheduled_report" {
		t.Errorf("expected type name 'jumpcloud_scheduled_report', got %q", resp.TypeName)
	}
}

func TestScheduledReportDataSource_Schema(t *testing.T) {
	d := NewScheduledReportDataSource()
	req := datasource.SchemaRequest{}
	var resp datasource.SchemaResponse
	d.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	optionalComputed := []string{"id", "display_name"}
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
		"description", "schedule_frequency", "schedule_time",
		"time_zone", "days", "is_active", "export_type",
		"report_id", "report_type", "recipients",
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
