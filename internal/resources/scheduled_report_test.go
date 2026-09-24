// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

func TestScheduledReportResource_Metadata(t *testing.T) {
	r := NewScheduledReportResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_scheduled_report" {
		t.Errorf("expected type name 'jumpcloud_scheduled_report', got %q", resp.TypeName)
	}
}

func TestScheduledReportResource_Schema(t *testing.T) {
	r := NewScheduledReportResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes

	required := []string{"display_name", "schedule_frequency"}
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

	optional := []string{
		"description", "schedule_time", "time_zone", "days",
		"start_date", "end_date", "is_active", "export_type",
		"report_id", "report_type", "recipients",
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

	if attr, ok := attrs["id"]; !ok {
		t.Error("expected attribute 'id' to exist")
	} else if !attr.IsComputed() {
		t.Error("expected attribute 'id' to be computed")
	}
}

func TestScheduledReportResource_ModelToAPI(t *testing.T) {
	r := &ScheduledReportResource{}
	ctx := context.Background()

	days, _ := types.ListValueFrom(ctx, types.StringType, []string{"MONDAY", "FRIDAY"})
	recipients, _ := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: recipientAttrTypes(),
	}, []ScheduledReportRecipientModel{
		{
			ChannelType: types.StringValue("email"),
			Value:       mustStringList(ctx, []string{"admin@example.com"}),
		},
	})

	model := ScheduledReportResourceModel{
		DisplayName:       types.StringValue("Weekly Report"),
		Description:       types.StringValue("A weekly report"),
		ScheduleFrequency: types.StringValue("weekly"),
		ScheduleTime:      types.StringValue("09:00 AM"),
		TimeZone:          types.StringValue("America/New_York"),
		Days:              days,
		IsActive:          types.BoolValue(true),
		ExportType:        types.StringValue("csv"),
		ReportID:          types.StringValue("rpt-tmpl-1"),
		ReportType:        types.StringValue("report_template"),
		Recipients:        recipients,
		StartDate:         types.StringNull(),
		EndDate:           types.StringNull(),
	}

	report := r.modelToAPI(ctx, &model)

	if report.DisplayName != "Weekly Report" {
		t.Errorf("expected display name 'Weekly Report', got %q", report.DisplayName)
	}
	if report.ScheduleFrequency != "weekly" {
		t.Errorf("expected frequency 'weekly', got %q", report.ScheduleFrequency)
	}
	if report.ScheduleTime != "09:00 AM" {
		t.Errorf("expected schedule_time '09:00 AM', got %q", report.ScheduleTime)
	}
	if len(report.Days) != 2 || report.Days[0] != "MONDAY" {
		t.Errorf("expected days [MONDAY, FRIDAY], got %v", report.Days)
	}
	if !report.IsActive {
		t.Error("expected is_active to be true")
	}
	if report.ExportType != "csv" {
		t.Errorf("expected export_type 'csv', got %q", report.ExportType)
	}
	if len(report.Recipients) != 1 {
		t.Fatalf("expected 1 recipient, got %d", len(report.Recipients))
	}
	if report.Recipients[0].ChannelType != "email" {
		t.Errorf("expected channel_type 'email', got %q", report.Recipients[0].ChannelType)
	}
	if len(report.Recipients[0].Value) != 1 || report.Recipients[0].Value[0] != "admin@example.com" {
		t.Errorf("expected recipient value [admin@example.com], got %v", report.Recipients[0].Value)
	}
}

func TestScheduledReportResource_APIToModel(t *testing.T) {
	r := &ScheduledReportResource{}

	t.Run("full fields", func(t *testing.T) {
		report := &client.ScheduledReport{
			ID:                "rpt-123",
			DisplayName:       "Daily Compliance",
			Description:       "Daily compliance check",
			ScheduleFrequency: "daily",
			ScheduleTime:      "08:00 AM",
			TimeZone:          "UTC",
			IsActive:          true,
			ExportType:        "json",
			ReportID:          "view-1",
			ReportType:        "report_view",
			Recipients: []client.ScheduledReportRecipient{
				{ChannelType: "slack", Value: []string{"#compliance"}},
			},
		}

		var model ScheduledReportResourceModel
		r.apiToModel(report, &model)

		if model.ID.ValueString() != "rpt-123" {
			t.Errorf("expected id 'rpt-123', got %q", model.ID.ValueString())
		}
		if model.DisplayName.ValueString() != "Daily Compliance" {
			t.Errorf("expected display name 'Daily Compliance', got %q", model.DisplayName.ValueString())
		}
		if model.ScheduleFrequency.ValueString() != "daily" {
			t.Errorf("expected frequency 'daily', got %q", model.ScheduleFrequency.ValueString())
		}
		if !model.IsActive.ValueBool() {
			t.Error("expected is_active to be true")
		}
		if model.Recipients.IsNull() {
			t.Error("expected recipients to not be null")
		}
	})

	t.Run("minimal fields", func(t *testing.T) {
		report := &client.ScheduledReport{
			ID:                "rpt-456",
			DisplayName:       "Minimal",
			ScheduleFrequency: "monthly",
		}

		var model ScheduledReportResourceModel
		r.apiToModel(report, &model)

		if !model.Description.IsNull() {
			t.Errorf("expected description to be null, got %q", model.Description.ValueString())
		}
		if !model.ScheduleTime.IsNull() {
			t.Errorf("expected schedule_time to be null, got %q", model.ScheduleTime.ValueString())
		}
		if !model.Days.IsNull() {
			t.Error("expected days to be null")
		}
		if model.Recipients.IsNull() != true {
			t.Error("expected recipients to be null")
		}
	})
}

func recipientAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"channel_type": types.StringType,
		"value":        types.ListType{ElemType: types.StringType},
	}
}

func mustStringList(ctx context.Context, vals []string) types.List {
	l, _ := types.ListValueFrom(ctx, types.StringType, vals)
	return l
}
