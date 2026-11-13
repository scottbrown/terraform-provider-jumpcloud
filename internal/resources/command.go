// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &CommandResource{}
var _ resource.ResourceWithImportState = &CommandResource{}

type CommandResource struct {
	client *client.Client
}

type CommandResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Command        types.String `tfsdk:"command"`
	CommandType    types.String `tfsdk:"command_type"`
	User           types.String `tfsdk:"user"`
	Sudo           types.Bool   `tfsdk:"sudo"`
	Shell          types.String `tfsdk:"shell"`
	Trigger        types.String `tfsdk:"trigger"`
	Schedule       types.String `tfsdk:"schedule"`
	ScheduleRepeat types.String `tfsdk:"schedule_repeat"`
	Timeout        types.String `tfsdk:"timeout"`
	LaunchType     types.String `tfsdk:"launch_type"`
}

func NewCommandResource() resource.Resource {
	return &CommandResource{}
}

func (r *CommandResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_command"
}

func (r *CommandResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud command (script that can be executed on devices).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the command.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the command.",
				Required:    true,
			},
			"command": schema.StringAttribute{
				Description: "The script or command to execute.",
				Required:    true,
			},
			"command_type": schema.StringAttribute{
				Description: "The platform type: linux, mac, or windows.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Description: "The user context to execute the command as. Defaults to root.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("root"),
			},
			"sudo": schema.BoolAttribute{
				Description: "Whether to run with elevated privileges.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"shell": schema.StringAttribute{
				Description: "Shell to use: bash, sh, zsh, or powershell.",
				Optional:    true,
			},
			"trigger": schema.StringAttribute{
				Description: "Trigger type: manual, automatic, deadline, or periodic.",
				Optional:    true,
			},
			"schedule": schema.StringAttribute{
				Description: "Cron schedule expression for periodic commands.",
				Optional:    true,
			},
			"schedule_repeat": schema.StringAttribute{
				Description: "Schedule repeat type: once, daily, weekly, or monthly.",
				Optional:    true,
			},
			"timeout": schema.StringAttribute{
				Description: "Command timeout duration.",
				Optional:    true,
			},
			"launch_type": schema.StringAttribute{
				Description: "Launch type for the command.",
				Optional:    true,
			},
		},
	}
}

func (r *CommandResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CommandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CommandResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmd := r.modelToAPI(&plan)
	created, err := r.client.CreateCommand(ctx, cmd)
	if err != nil {
		resp.Diagnostics.AddError("Error creating command", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CommandResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CommandResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmd, err := r.client.GetCommand(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading command", err.Error())
		return
	}

	r.apiToModel(cmd, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CommandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CommandResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CommandResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cmd := r.modelToAPI(&plan)
	_, err := r.client.UpdateCommand(ctx, state.ID.ValueString(), cmd)
	if err != nil {
		resp.Diagnostics.AddError("Error updating command", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CommandResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CommandResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCommand(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting command", err.Error())
		return
	}
}

func (r *CommandResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	cmd, err := r.client.GetCommand(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing command", fmt.Sprintf("Cannot read command %s: %s", req.ID, err))
		return
	}

	var state CommandResourceModel
	state.ID = types.StringValue(cmd.ID)
	r.apiToModel(cmd, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CommandResource) modelToAPI(model *CommandResourceModel) *client.Command {
	return &client.Command{
		Name:           model.Name.ValueString(),
		Command:        strings.TrimSuffix(model.Command.ValueString(), "\n"),
		CommandType:    model.CommandType.ValueString(),
		User:           model.User.ValueString(),
		Sudo:           model.Sudo.ValueBool(),
		Shell:          model.Shell.ValueString(),
		Trigger:        model.Trigger.ValueString(),
		Schedule:       model.Schedule.ValueString(),
		ScheduleRepeat: model.ScheduleRepeat.ValueString(),
		Timeout:        model.Timeout.ValueString(),
		LaunchType:     model.LaunchType.ValueString(),
	}
}

func (r *CommandResource) apiToModel(cmd *client.Command, model *CommandResourceModel) {
	model.Name = types.StringValue(cmd.Name)
	command := cmd.Command
	if !strings.HasSuffix(command, "\n") {
		command += "\n"
	}
	model.Command = types.StringValue(command)
	model.CommandType = types.StringValue(cmd.CommandType)
	model.User = types.StringValue(cmd.User)
	model.Sudo = types.BoolValue(cmd.Sudo)
	if cmd.Shell != "" {
		model.Shell = types.StringValue(cmd.Shell)
	}
	if cmd.Trigger != "" {
		model.Trigger = types.StringValue(cmd.Trigger)
	}
	if cmd.Schedule != "" {
		model.Schedule = types.StringValue(cmd.Schedule)
	}
	if cmd.ScheduleRepeat != "" {
		model.ScheduleRepeat = types.StringValue(cmd.ScheduleRepeat)
	}
	if cmd.Timeout != "" {
		model.Timeout = types.StringValue(cmd.Timeout)
	}
	if cmd.LaunchType != "" {
		model.LaunchType = types.StringValue(cmd.LaunchType)
	}
}
