// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ datasource.DataSource = &RoleDataSource{}

type RoleDataSource struct {
	client *client.Client
}

type RoleDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.List   `tfsdk:"permissions"`
	Scope       types.Object `tfsdk:"scope"`
}

func NewRoleDataSource() datasource.DataSource {
	return &RoleDataSource{}
}

func (d *RoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *RoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud administrative role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the role.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the role.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the role.",
				Computed:    true,
			},
			"permissions": schema.ListNestedAttribute{
				Description: "List of permission grants for this role.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource_type": schema.StringAttribute{
							Description: "The resource type this permission applies to.",
							Computed:    true,
						},
						"action": schema.StringAttribute{
							Description: "The action allowed on the resource type.",
							Computed:    true,
						},
					},
				},
			},
			"scope": schema.SingleNestedAttribute{
				Description: "Scope restriction for the role.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Scope type: all or scope_group.",
						Computed:    true,
					},
					"id": schema.StringAttribute{
						Description: "The scope group ID.",
						Computed:    true,
					},
				},
			},
		},
	}
}

func (d *RoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RoleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	role, err := d.client.GetRole(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading role", err.Error())
		return
	}

	config.Name = types.StringValue(role.Name)
	if role.Description != "" {
		config.Description = types.StringValue(role.Description)
	} else {
		config.Description = types.StringNull()
	}

	if len(role.Permissions) > 0 {
		permObjType := types.ObjectType{AttrTypes: dsRolePermissionAttrTypes()}
		permValues := make([]attr.Value, len(role.Permissions))
		for i, p := range role.Permissions {
			permValues[i], _ = types.ObjectValue(dsRolePermissionAttrTypes(), map[string]attr.Value{
				"resource_type": types.StringValue(p.ResourceType),
				"action":        types.StringValue(p.Action),
			})
		}
		config.Permissions, _ = types.ListValue(permObjType, permValues)
	} else {
		config.Permissions, _ = types.ListValue(types.ObjectType{AttrTypes: dsRolePermissionAttrTypes()}, []attr.Value{})
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
		config.Scope, _ = types.ObjectValue(dsRoleScopeAttrTypes(), scopeAttrs)
	} else {
		config.Scope = types.ObjectNull(dsRoleScopeAttrTypes())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func dsRolePermissionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"resource_type": types.StringType,
		"action":        types.StringType,
	}
}

func dsRoleScopeAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"id":   types.StringType,
	}
}
