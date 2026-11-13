// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &IPListResource{}
var _ resource.ResourceWithImportState = &IPListResource{}

type IPListResource struct {
	client *client.Client
}

type IPListResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IPs         types.List   `tfsdk:"ips"`
}

func NewIPListResource() resource.Resource {
	return &IPListResource{}
}

func (r *IPListResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_list"
}

func (r *IPListResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud IP list for conditional access policies.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the IP list.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the IP list.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the IP list.",
				Optional:    true,
			},
			"ips": schema.ListAttribute{
				Description: "List of IP addresses or CIDR ranges.",
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *IPListResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IPListResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IPListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipList := r.modelToAPI(ctx, &plan)
	created, err := r.client.CreateIPList(ctx, ipList)
	if err != nil {
		resp.Diagnostics.AddError("Error creating IP list", err.Error())
		return
	}

	r.apiToModel(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IPListResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IPListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipList, err := r.client.GetIPList(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading IP list", err.Error())
		return
	}

	r.apiToModel(ipList, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IPListResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IPListResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state IPListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ipList := r.modelToAPI(ctx, &plan)
	updated, err := r.client.UpdateIPList(ctx, state.ID.ValueString(), ipList)
	if err != nil {
		resp.Diagnostics.AddError("Error updating IP list", err.Error())
		return
	}

	r.apiToModel(updated, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IPListResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IPListResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteIPList(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting IP list", err.Error())
		return
	}
}

func (r *IPListResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ipList, err := r.client.GetIPList(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing IP list", fmt.Sprintf("Cannot read IP list %s: %s", req.ID, err))
		return
	}

	var state IPListResourceModel
	r.apiToModel(ipList, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IPListResource) modelToAPI(ctx context.Context, model *IPListResourceModel) *client.IPList {
	ipList := &client.IPList{
		Name: model.Name.ValueString(),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		ipList.Description = model.Description.ValueString()
	}

	if !model.IPs.IsNull() && !model.IPs.IsUnknown() {
		var ips []string
		model.IPs.ElementsAs(ctx, &ips, false)
		ipList.IPs = ips
	}

	return ipList
}

func (r *IPListResource) apiToModel(ipList *client.IPList, model *IPListResourceModel) {
	model.ID = types.StringValue(ipList.ID)
	model.Name = types.StringValue(ipList.Name)

	if ipList.Description != "" {
		model.Description = types.StringValue(ipList.Description)
	} else {
		model.Description = types.StringNull()
	}

	if len(ipList.IPs) > 0 {
		ipValues := make([]attr.Value, len(ipList.IPs))
		for i, ip := range ipList.IPs {
			ipValues[i] = types.StringValue(ip)
		}
		model.IPs, _ = types.ListValue(types.StringType, ipValues)
	} else {
		model.IPs, _ = types.ListValue(types.StringType, []attr.Value{})
	}
}
