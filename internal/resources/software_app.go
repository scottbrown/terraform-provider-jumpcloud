// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &SoftwareAppResource{}
var _ resource.ResourceWithImportState = &SoftwareAppResource{}

type SoftwareAppResource struct {
	client *client.Client
}

type SoftwareAppResourceModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Settings    types.String `tfsdk:"settings"`
}

func NewSoftwareAppResource() resource.Resource {
	return &SoftwareAppResource{}
}

func (r *SoftwareAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_software_app"
}

func (r *SoftwareAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud software application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the software application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the software application.",
				Required:    true,
			},
			"settings": schema.StringAttribute{
				Description: "JSON-encoded settings for the software application. Structure depends on the application type.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *SoftwareAppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SoftwareAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SoftwareAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app := r.modelToAPI(&plan)
	created, err := r.client.CreateSoftwareApp(ctx, app)
	if err != nil {
		resp.Diagnostics.AddError("Error creating software app", err.Error())
		return
	}

	r.apiToModel(created, &plan)
	plan.ID = types.StringValue(created.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SoftwareAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SoftwareAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetSoftwareApp(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading software app", err.Error())
		return
	}

	r.apiToModel(app, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SoftwareAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SoftwareAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SoftwareAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app := r.modelToAPI(&plan)
	updated, err := r.client.UpdateSoftwareApp(ctx, state.ID.ValueString(), app)
	if err != nil {
		resp.Diagnostics.AddError("Error updating software app", err.Error())
		return
	}

	r.apiToModel(updated, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SoftwareAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SoftwareAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSoftwareApp(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting software app", err.Error())
		return
	}
}

func (r *SoftwareAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	app, err := r.client.GetSoftwareApp(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing software app", fmt.Sprintf("Cannot read software app %s: %s", req.ID, err))
		return
	}

	var state SoftwareAppResourceModel
	state.ID = types.StringValue(app.ID)
	r.apiToModel(app, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SoftwareAppResource) modelToAPI(model *SoftwareAppResourceModel) *client.SoftwareApp {
	app := &client.SoftwareApp{
		DisplayName: model.DisplayName.ValueString(),
	}
	if !model.Settings.IsNull() && !model.Settings.IsUnknown() {
		app.Settings = json.RawMessage(model.Settings.ValueString())
	}
	return app
}

func (r *SoftwareAppResource) apiToModel(app *client.SoftwareApp, model *SoftwareAppResourceModel) {
	model.ID = types.StringValue(app.ID)
	model.DisplayName = types.StringValue(app.DisplayName)
	if len(app.Settings) > 0 && string(app.Settings) != "null" {
		var normalized any
		if err := json.Unmarshal(app.Settings, &normalized); err == nil {
			compact, _ := json.Marshal(normalized)
			model.Settings = types.StringValue(string(compact))
		} else {
			model.Settings = types.StringValue(string(app.Settings))
		}
	} else {
		model.Settings = types.StringNull()
	}
}
