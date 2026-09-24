// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &ScheduledReportResource{}
var _ resource.ResourceWithImportState = &ScheduledReportResource{}

type ScheduledReportResource struct {
	client *client.Client
}

type ScheduledReportResourceModel struct {
	ID                types.String `tfsdk:"id"`
	DisplayName       types.String `tfsdk:"display_name"`
	Description       types.String `tfsdk:"description"`
	ScheduleFrequency types.String `tfsdk:"schedule_frequency"`
	ScheduleTime      types.String `tfsdk:"schedule_time"`
	TimeZone          types.String `tfsdk:"time_zone"`
	Days              types.List   `tfsdk:"days"`
	StartDate         types.String `tfsdk:"start_date"`
	EndDate           types.String `tfsdk:"end_date"`
	IsActive          types.Bool   `tfsdk:"is_active"`
	ExportType        types.String `tfsdk:"export_type"`
	ReportID          types.String `tfsdk:"report_id"`
	ReportType        types.String `tfsdk:"report_type"`
	Recipients        types.List   `tfsdk:"recipients"`
}

type ScheduledReportRecipientModel struct {
	ChannelType types.String `tfsdk:"channel_type"`
	Value       types.List   `tfsdk:"value"`
}

func NewScheduledReportResource() resource.Resource {
	return &ScheduledReportResource{}
}

func (r *ScheduledReportResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scheduled_report"
}

func (r *ScheduledReportResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud scheduled report.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the scheduled report.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the scheduled report.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the scheduled report.",
				Optional:    true,
			},
			"schedule_frequency": schema.StringAttribute{
				Description: "How often the report runs: daily, weekly, monthly, or quarterly.",
				Required:    true,
			},
			"schedule_time": schema.StringAttribute{
				Description: "Time of day to run the report (e.g. '09:00 AM').",
				Optional:    true,
			},
			"time_zone": schema.StringAttribute{
				Description: "Time zone for the schedule (e.g. 'America/New_York').",
				Optional:    true,
			},
			"days": schema.ListAttribute{
				Description: "Days to run the report. For weekly: day names (MONDAY, TUESDAY, etc). For monthly/quarterly: FIRST_DAY or LAST_DAY.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"start_date": schema.StringAttribute{
				Description: "Start date for the schedule (ISO 8601 format).",
				Optional:    true,
			},
			"end_date": schema.StringAttribute{
				Description: "End date for the schedule (ISO 8601 format).",
				Optional:    true,
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether the scheduled report is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"export_type": schema.StringAttribute{
				Description: "Export format: csv or json.",
				Optional:    true,
			},
			"report_id": schema.StringAttribute{
				Description: "The ID of the report view or template to run.",
				Optional:    true,
			},
			"report_type": schema.StringAttribute{
				Description: "Type of report: report_view or report_template.",
				Optional:    true,
			},
			"recipients": schema.ListNestedAttribute{
				Description: "Notification recipients for the report.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"channel_type": schema.StringAttribute{
							Description: "Channel type: email, slack, or webhook.",
							Required:    true,
						},
						"value": schema.ListAttribute{
							Description: "List of recipient addresses (emails, Slack channels, or webhook URLs).",
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (r *ScheduledReportResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *client.Client")
		return
	}
	r.client = c
}

func (r *ScheduledReportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ScheduledReportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	report := r.modelToAPI(ctx, &plan)
	created, err := r.client.CreateScheduledReport(ctx, report)
	if err != nil {
		resp.Diagnostics.AddError("Error creating scheduled report", err.Error())
		return
	}

	r.apiToModel(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ScheduledReportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ScheduledReportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	report, err := r.client.GetScheduledReport(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading scheduled report", err.Error())
		return
	}

	r.apiToModel(report, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ScheduledReportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ScheduledReportResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ScheduledReportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	report := r.modelToAPI(ctx, &plan)
	updated, err := r.client.UpdateScheduledReport(ctx, state.ID.ValueString(), report)
	if err != nil {
		resp.Diagnostics.AddError("Error updating scheduled report", err.Error())
		return
	}

	r.apiToModel(updated, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ScheduledReportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ScheduledReportResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteScheduledReport(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting scheduled report", err.Error())
		return
	}
}

func (r *ScheduledReportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	report, err := r.client.GetScheduledReport(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing scheduled report", fmt.Sprintf("Cannot read scheduled report %s: %s", req.ID, err))
		return
	}

	var state ScheduledReportResourceModel
	r.apiToModel(report, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ScheduledReportResource) modelToAPI(ctx context.Context, model *ScheduledReportResourceModel) *client.ScheduledReport {
	report := &client.ScheduledReport{
		DisplayName:       model.DisplayName.ValueString(),
		ScheduleFrequency: model.ScheduleFrequency.ValueString(),
		IsActive:          model.IsActive.ValueBool(),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		report.Description = model.Description.ValueString()
	}
	if !model.ScheduleTime.IsNull() && !model.ScheduleTime.IsUnknown() {
		report.ScheduleTime = model.ScheduleTime.ValueString()
	}
	if !model.TimeZone.IsNull() && !model.TimeZone.IsUnknown() {
		report.TimeZone = model.TimeZone.ValueString()
	}
	if !model.StartDate.IsNull() && !model.StartDate.IsUnknown() {
		report.StartDate = model.StartDate.ValueString()
	}
	if !model.EndDate.IsNull() && !model.EndDate.IsUnknown() {
		report.EndDate = model.EndDate.ValueString()
	}
	if !model.ExportType.IsNull() && !model.ExportType.IsUnknown() {
		report.ExportType = model.ExportType.ValueString()
	}
	if !model.ReportID.IsNull() && !model.ReportID.IsUnknown() {
		report.ReportID = model.ReportID.ValueString()
	}
	if !model.ReportType.IsNull() && !model.ReportType.IsUnknown() {
		report.ReportType = model.ReportType.ValueString()
	}

	if !model.Days.IsNull() && !model.Days.IsUnknown() {
		var days []string
		model.Days.ElementsAs(ctx, &days, false)
		report.Days = days
	}

	if !model.Recipients.IsNull() && !model.Recipients.IsUnknown() {
		var recipientModels []ScheduledReportRecipientModel
		model.Recipients.ElementsAs(ctx, &recipientModels, false)
		for _, rm := range recipientModels {
			var values []string
			rm.Value.ElementsAs(ctx, &values, false)
			report.Recipients = append(report.Recipients, client.ScheduledReportRecipient{
				ChannelType: rm.ChannelType.ValueString(),
				Value:       values,
			})
		}
	}

	return report
}

func (r *ScheduledReportResource) apiToModel(report *client.ScheduledReport, model *ScheduledReportResourceModel) {
	model.ID = types.StringValue(report.ID)
	model.DisplayName = types.StringValue(report.DisplayName)
	model.ScheduleFrequency = types.StringValue(report.ScheduleFrequency)
	model.IsActive = types.BoolValue(report.IsActive)

	if report.Description != "" {
		model.Description = types.StringValue(report.Description)
	} else {
		model.Description = types.StringNull()
	}
	if report.ScheduleTime != "" {
		model.ScheduleTime = types.StringValue(report.ScheduleTime)
	} else {
		model.ScheduleTime = types.StringNull()
	}
	if report.TimeZone != "" {
		model.TimeZone = types.StringValue(report.TimeZone)
	} else {
		model.TimeZone = types.StringNull()
	}
	if report.StartDate != "" {
		model.StartDate = types.StringValue(report.StartDate)
	} else {
		model.StartDate = types.StringNull()
	}
	if report.EndDate != "" {
		model.EndDate = types.StringValue(report.EndDate)
	} else {
		model.EndDate = types.StringNull()
	}
	if report.ExportType != "" {
		model.ExportType = types.StringValue(report.ExportType)
	} else {
		model.ExportType = types.StringNull()
	}
	if report.ReportID != "" {
		model.ReportID = types.StringValue(report.ReportID)
	} else {
		model.ReportID = types.StringNull()
	}
	if report.ReportType != "" {
		model.ReportType = types.StringValue(report.ReportType)
	} else {
		model.ReportType = types.StringNull()
	}

	if len(report.Days) > 0 {
		dayValues := make([]attr.Value, len(report.Days))
		for i, d := range report.Days {
			dayValues[i] = types.StringValue(d)
		}
		model.Days, _ = types.ListValue(types.StringType, dayValues)
	} else {
		model.Days = types.ListNull(types.StringType)
	}

	if len(report.Recipients) > 0 {
		recipientType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"channel_type": types.StringType,
				"value":        types.ListType{ElemType: types.StringType},
			},
		}

		recipientValues := make([]attr.Value, len(report.Recipients))
		for i, rec := range report.Recipients {
			valList := make([]attr.Value, len(rec.Value))
			for j, v := range rec.Value {
				valList[j] = types.StringValue(v)
			}
			valuesList, _ := types.ListValue(types.StringType, valList)
			recipientValues[i], _ = types.ObjectValue(
				recipientType.AttrTypes,
				map[string]attr.Value{
					"channel_type": types.StringValue(rec.ChannelType),
					"value":        valuesList,
				},
			)
		}
		model.Recipients, _ = types.ListValue(recipientType, recipientValues)
	} else {
		model.Recipients = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"channel_type": types.StringType,
				"value":        types.ListType{ElemType: types.StringType},
			},
		})
	}
}
