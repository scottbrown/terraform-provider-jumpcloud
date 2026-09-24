// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package datasources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ datasource.DataSource = &ScheduledReportDataSource{}

type ScheduledReportDataSource struct {
	client *client.Client
}

type ScheduledReportDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	DisplayName       types.String `tfsdk:"display_name"`
	Description       types.String `tfsdk:"description"`
	ScheduleFrequency types.String `tfsdk:"schedule_frequency"`
	ScheduleTime      types.String `tfsdk:"schedule_time"`
	TimeZone          types.String `tfsdk:"time_zone"`
	Days              types.List   `tfsdk:"days"`
	IsActive          types.Bool   `tfsdk:"is_active"`
	ExportType        types.String `tfsdk:"export_type"`
	ReportID          types.String `tfsdk:"report_id"`
	ReportType        types.String `tfsdk:"report_type"`
	Recipients        types.String `tfsdk:"recipients"`
}

func NewScheduledReportDataSource() datasource.DataSource {
	return &ScheduledReportDataSource{}
}

func (d *ScheduledReportDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scheduled_report"
}

func (d *ScheduledReportDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud scheduled report by ID or display name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the scheduled report. Exactly one of id or display_name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the scheduled report. Exactly one of id or display_name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the scheduled report.",
				Computed:    true,
			},
			"schedule_frequency": schema.StringAttribute{
				Description: "How often the report runs: daily, weekly, monthly, or quarterly.",
				Computed:    true,
			},
			"schedule_time": schema.StringAttribute{
				Description: "Time of day to run the report.",
				Computed:    true,
			},
			"time_zone": schema.StringAttribute{
				Description: "Time zone for the schedule.",
				Computed:    true,
			},
			"days": schema.ListAttribute{
				Description: "Days to run the report.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the scheduled report is active.",
				Computed:    true,
			},
			"export_type": schema.StringAttribute{
				Description: "Export format: csv or json.",
				Computed:    true,
			},
			"report_id": schema.StringAttribute{
				Description: "The ID of the report view or template.",
				Computed:    true,
			},
			"report_type": schema.StringAttribute{
				Description: "Type of report: report_view or report_template.",
				Computed:    true,
			},
			"recipients": schema.StringAttribute{
				Description: "JSON-encoded list of notification recipients.",
				Computed:    true,
			},
		},
	}
}

func (d *ScheduledReportDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected DataSource Configure Type", "Expected *client.Client")
		return
	}
	d.client = c
}

func (d *ScheduledReportDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ScheduledReportDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && !config.ID.IsUnknown()
	hasName := !config.DisplayName.IsNull() && !config.DisplayName.IsUnknown()

	if !hasID && !hasName {
		resp.Diagnostics.AddError("Missing attribute", "Exactly one of 'id' or 'display_name' must be specified.")
		return
	}
	if hasID && hasName {
		resp.Diagnostics.AddError("Conflicting attributes", "Only one of 'id' or 'display_name' may be specified.")
		return
	}

	var report *client.ScheduledReport
	var err error

	if hasID {
		report, err = d.client.GetScheduledReport(ctx, config.ID.ValueString())
	} else {
		report, err = d.client.GetScheduledReportByName(ctx, config.DisplayName.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Error reading scheduled report", err.Error())
		return
	}

	config.ID = types.StringValue(report.ID)
	config.DisplayName = types.StringValue(report.DisplayName)
	config.ScheduleFrequency = types.StringValue(report.ScheduleFrequency)
	config.IsActive = types.BoolValue(report.IsActive)

	if report.Description != "" {
		config.Description = types.StringValue(report.Description)
	} else {
		config.Description = types.StringNull()
	}
	if report.ScheduleTime != "" {
		config.ScheduleTime = types.StringValue(report.ScheduleTime)
	} else {
		config.ScheduleTime = types.StringNull()
	}
	if report.TimeZone != "" {
		config.TimeZone = types.StringValue(report.TimeZone)
	} else {
		config.TimeZone = types.StringNull()
	}
	if report.ExportType != "" {
		config.ExportType = types.StringValue(report.ExportType)
	} else {
		config.ExportType = types.StringNull()
	}
	if report.ReportID != "" {
		config.ReportID = types.StringValue(report.ReportID)
	} else {
		config.ReportID = types.StringNull()
	}
	if report.ReportType != "" {
		config.ReportType = types.StringValue(report.ReportType)
	} else {
		config.ReportType = types.StringNull()
	}

	if len(report.Days) > 0 {
		dayValues := make([]attr.Value, len(report.Days))
		for i, d := range report.Days {
			dayValues[i] = types.StringValue(d)
		}
		config.Days, _ = types.ListValue(types.StringType, dayValues)
	} else {
		config.Days = types.ListNull(types.StringType)
	}

	if len(report.Recipients) > 0 {
		recipientsJSON, _ := json.Marshal(report.Recipients)
		config.Recipients = types.StringValue(string(recipientsJSON))
	} else {
		config.Recipients = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
