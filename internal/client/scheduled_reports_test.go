// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateScheduledReport(t *testing.T) {
	var gotBody scheduledReportRequest
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/reports/scheduled" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := scheduledReportResponse{
			ScheduledReport: *gotBody.ScheduledReport,
		}
		resp.ScheduledReport.ID = "new-report-id"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	report, err := c.CreateScheduledReport(context.Background(), &ScheduledReport{
		DisplayName:       "Daily User Report",
		ScheduleFrequency: "daily",
		ScheduleTime:      "09:00 AM",
		TimeZone:          "America/New_York",
		IsActive:          true,
		ExportType:        "csv",
		Recipients: []ScheduledReportRecipient{
			{ChannelType: "email", Value: []string{"admin@example.com"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ID != "new-report-id" {
		t.Errorf("expected id 'new-report-id', got %q", report.ID)
	}
	if gotBody.ScheduledReport.DisplayName != "Daily User Report" {
		t.Errorf("expected display name 'Daily User Report', got %q", gotBody.ScheduledReport.DisplayName)
	}
	if gotBody.ScheduledReport.ScheduleFrequency != "daily" {
		t.Errorf("expected frequency 'daily', got %q", gotBody.ScheduledReport.ScheduleFrequency)
	}
}

func TestGetScheduledReport(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/reports/scheduled/rpt-123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(scheduledReportResponse{
			ScheduledReport: ScheduledReport{
				ID:                "rpt-123",
				DisplayName:       "Weekly Compliance",
				ScheduleFrequency: "weekly",
				Days:              []string{"MONDAY"},
			},
		})
	}))

	report, err := c.GetScheduledReport(context.Background(), "rpt-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ID != "rpt-123" {
		t.Errorf("expected id 'rpt-123', got %q", report.ID)
	}
	if report.DisplayName != "Weekly Compliance" {
		t.Errorf("expected display name 'Weekly Compliance', got %q", report.DisplayName)
	}
	if len(report.Days) != 1 || report.Days[0] != "MONDAY" {
		t.Errorf("expected days [MONDAY], got %v", report.Days)
	}
}

func TestUpdateScheduledReport(t *testing.T) {
	var gotBody scheduledReportRequest
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/reports/scheduled/rpt-123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := scheduledReportResponse{
			ScheduledReport: *gotBody.ScheduledReport,
		}
		resp.ScheduledReport.ID = "rpt-123"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	report, err := c.UpdateScheduledReport(context.Background(), "rpt-123", &ScheduledReport{
		DisplayName:       "Updated Report",
		ScheduleFrequency: "monthly",
		Days:              []string{"FIRST_DAY"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.DisplayName != "Updated Report" {
		t.Errorf("expected display name 'Updated Report', got %q", report.DisplayName)
	}
}

func TestDeleteScheduledReport(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/reports/scheduled/rpt-del" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))

	err := c.DeleteScheduledReport(context.Background(), "rpt-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListScheduledReports(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/reports/scheduled" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(scheduledReportListResponse{
			ScheduledReports: []ScheduledReport{
				{ID: "1", DisplayName: "Report A"},
				{ID: "2", DisplayName: "Report B"},
			},
		})
	}))

	reports, err := c.ListScheduledReports(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reports) != 2 {
		t.Errorf("expected 2 reports, got %d", len(reports))
	}
}

func TestGetScheduledReportByName(t *testing.T) {
	callCount := 0
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			_ = json.NewEncoder(w).Encode(scheduledReportListResponse{
				ScheduledReports: []ScheduledReport{
					{ID: "1", DisplayName: "Report A"},
					{ID: "2", DisplayName: "Target Report"},
				},
			})
		} else {
			_ = json.NewEncoder(w).Encode(scheduledReportResponse{
				ScheduledReport: ScheduledReport{
					ID:                "2",
					DisplayName:       "Target Report",
					ScheduleFrequency: "daily",
				},
			})
		}
	}))

	report, err := c.GetScheduledReportByName(context.Background(), "Target Report")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.ID != "2" {
		t.Errorf("expected id '2', got %q", report.ID)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (list + get), got %d", callCount)
	}
}

func TestGetScheduledReportByName_NotFound(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(scheduledReportListResponse{
			ScheduledReports: []ScheduledReport{},
		})
	}))

	_, err := c.GetScheduledReportByName(context.Background(), "Missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
