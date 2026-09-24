// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"fmt"
)

type ScheduledReport struct {
	ID                string                     `json:"id,omitempty"`
	DisplayName       string                     `json:"displayName"`
	Description       string                     `json:"description,omitempty"`
	ScheduleFrequency string                     `json:"scheduleFrequency"`
	ScheduleTime      string                     `json:"scheduleTime,omitempty"`
	TimeZone          string                     `json:"timeZone,omitempty"`
	Days              []string                   `json:"days,omitempty"`
	StartDate         string                     `json:"startDate,omitempty"`
	EndDate           string                     `json:"endDate,omitempty"`
	IsActive          bool                       `json:"isActive,omitempty"`
	ExportType        string                     `json:"exportType,omitempty"`
	ReportID          string                     `json:"reportId,omitempty"`
	ReportType        string                     `json:"reportType,omitempty"`
	Recipients        []ScheduledReportRecipient `json:"recipients,omitempty"`
}

type ScheduledReportRecipient struct {
	ChannelType string   `json:"channelType"`
	Value       []string `json:"value"`
}

type scheduledReportRequest struct {
	ScheduledReport *ScheduledReport `json:"scheduledReport"`
}

type scheduledReportResponse struct {
	ScheduledReport ScheduledReport `json:"scheduledReport"`
}

type scheduledReportListResponse struct {
	ScheduledReports []ScheduledReport `json:"scheduledReports"`
}

func (c *Client) CreateScheduledReport(ctx context.Context, report *ScheduledReport) (*ScheduledReport, error) {
	var resp scheduledReportResponse
	body := scheduledReportRequest{ScheduledReport: report}
	err := c.do(ctx, "POST", "/reports/scheduled", body, &resp)
	if err != nil {
		return nil, fmt.Errorf("creating scheduled report: %w", err)
	}
	return &resp.ScheduledReport, nil
}

func (c *Client) GetScheduledReport(ctx context.Context, id string) (*ScheduledReport, error) {
	var resp scheduledReportResponse
	err := c.do(ctx, "GET", "/reports/scheduled/"+id, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("getting scheduled report: %w", err)
	}
	return &resp.ScheduledReport, nil
}

func (c *Client) UpdateScheduledReport(ctx context.Context, id string, report *ScheduledReport) (*ScheduledReport, error) {
	var resp scheduledReportResponse
	body := scheduledReportRequest{ScheduledReport: report}
	err := c.do(ctx, "PUT", "/reports/scheduled/"+id, body, &resp)
	if err != nil {
		return nil, fmt.Errorf("updating scheduled report: %w", err)
	}
	return &resp.ScheduledReport, nil
}

func (c *Client) DeleteScheduledReport(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/reports/scheduled/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting scheduled report: %w", err)
	}
	return nil
}

func (c *Client) ListScheduledReports(ctx context.Context) ([]ScheduledReport, error) {
	var resp scheduledReportListResponse
	err := c.do(ctx, "GET", "/reports/scheduled", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("listing scheduled reports: %w", err)
	}
	return resp.ScheduledReports, nil
}

func (c *Client) GetScheduledReportByName(ctx context.Context, name string) (*ScheduledReport, error) {
	reports, err := c.ListScheduledReports(ctx)
	if err != nil {
		return nil, err
	}
	for i := range reports {
		if reports[i].DisplayName == name {
			return c.GetScheduledReport(ctx, reports[i].ID)
		}
	}
	return nil, fmt.Errorf("scheduled report with name %q not found", name)
}
