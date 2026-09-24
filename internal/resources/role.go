// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

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

var _ resource.Resource = &RoleResource{}
var _ resource.ResourceWithImportState = &RoleResource{}

type RoleResource struct {
	client *client.Client
}

type RoleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.List   `tfsdk:"permissions"`
	Scope       types.Object `tfsdk:"scope"`
}

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

func (r *RoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud administrative role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the role.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the role.",
				Optional:    true,
			},
			"permissions": schema.ListNestedAttribute{
				Description: "List of permission grants for this role.",
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource_type": schema.StringAttribute{
							Description: "The resource type this permission applies to.",
							Required:    true,
						},
						"action": schema.StringAttribute{
							Description: "The action allowed on the resource type.",
							Required:    true,
						},
					},
				},
			},
			"scope": schema.SingleNestedAttribute{
				Description: "Scope restriction for the role (all resources or a specific scope group).",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Scope type: all or scope_group.",
						Required:    true,
					},
					"id": schema.StringAttribute{
						Description: "The scope group ID (required when type is scope_group).",
						Optional:    true,
					},
				},
			},
		},
	}
}

func (r *RoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role := r.modelToAPI(ctx, &plan)
	created, err := r.client.CreateRole(ctx, role)
	if err != nil {
		resp.Diagnostics.AddError("Error creating role", err.Error())
		return
	}

	r.apiToModel(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := r.client.GetRole(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading role", err.Error())
		return
	}

	r.apiToModel(role, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role := r.modelToAPI(ctx, &plan)
	updated, err := r.client.UpdateRole(ctx, state.ID.ValueString(), role)
	if err != nil {
		resp.Diagnostics.AddError("Error updating role", err.Error())
		return
	}

	r.apiToModel(updated, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRole(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting role", err.Error())
		return
	}
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	role, err := r.client.GetRole(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing role", fmt.Sprintf("Cannot read role %s: %s", req.ID, err))
		return
	}

	var state RoleResourceModel
	r.apiToModel(role, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RoleResource) modelToAPI(ctx context.Context, model *RoleResourceModel) *client.Role {
	role := &client.Role{
		Name: model.Name.ValueString(),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		role.Description = model.Description.ValueString()
	}

	if !model.Permissions.IsNull() && !model.Permissions.IsUnknown() {
		type permModel struct {
			ResourceType string `tfsdk:"resource_type"`
			Action       string `tfsdk:"action"`
		}
		var perms []permModel
		model.Permissions.ElementsAs(ctx, &perms, false)
		for _, p := range perms {
			role.Permissions = append(role.Permissions, client.RolePermission{
				ResourceType: p.ResourceType,
				Action:       p.Action,
			})
		}
	}

	if !model.Scope.IsNull() && !model.Scope.IsUnknown() {
		attrs := model.Scope.Attributes()
		scope := &client.RoleScope{}
		if v, ok := attrString(attrs, "type"); ok {
			scope.Type = v
		}
		if v, ok := attrString(attrs, "id"); ok {
			scope.ID = v
		}
		role.Scope = scope
	}

	return role
}

func (r *RoleResource) apiToModel(role *client.Role, model *RoleResourceModel) {
	model.ID = types.StringValue(role.ID)
	model.Name = types.StringValue(role.Name)

	if role.Description != "" {
		model.Description = types.StringValue(role.Description)
	} else {
		model.Description = types.StringNull()
	}

	if len(role.Permissions) > 0 {
		permObjType := types.ObjectType{AttrTypes: rolePermissionAttrTypes()}
		permValues := make([]attr.Value, len(role.Permissions))
		for i, p := range role.Permissions {
			permValues[i], _ = types.ObjectValue(rolePermissionAttrTypes(), map[string]attr.Value{
				"resource_type": types.StringValue(p.ResourceType),
				"action":        types.StringValue(p.Action),
			})
		}
		model.Permissions, _ = types.ListValue(permObjType, permValues)
	} else {
		model.Permissions, _ = types.ListValue(types.ObjectType{AttrTypes: rolePermissionAttrTypes()}, []attr.Value{})
	}

	if role.Scope != nil {
		scopeAttrs := map[string]attr.Value{
			"type": types.StringValue(role.Scope.Type),
		}
		if role.Scope.ID != "" {
			scopeAttrs["id"] = types.StringValue(role.Scope.ID)
		} else {
			scopeAttrs["id"] = types.StringNull()
		}
		model.Scope, _ = types.ObjectValue(roleScopeAttrTypes(), scopeAttrs)
	} else {
		model.Scope = types.ObjectNull(roleScopeAttrTypes())
	}
}

func rolePermissionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"resource_type": types.StringType,
		"action":        types.StringType,
	}
}

func roleScopeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"id":   types.StringType,
	}
}
