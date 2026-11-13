// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ datasource.DataSource = &AuthenticationPolicyDataSource{}

type AuthenticationPolicyDataSource struct {
	client *client.Client
}

type AuthenticationPolicyDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Disabled   types.Bool   `tfsdk:"disabled"`
	Conditions types.Object `tfsdk:"conditions"`
	Effect     types.Object `tfsdk:"effect"`
	Targets    types.List   `tfsdk:"targets"`
}

func NewAuthenticationPolicyDataSource() datasource.DataSource {
	return &AuthenticationPolicyDataSource{}
}

func (d *AuthenticationPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authentication_policy"
}

func (d *AuthenticationPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud authentication policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the authentication policy.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the authentication policy.",
				Computed:    true,
			},
			"type": schema.StringAttribute{
				Description: "The policy type.",
				Computed:    true,
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the policy is disabled.",
				Computed:    true,
			},
			"conditions": schema.SingleNestedAttribute{
				Description: "Conditions that determine when this policy applies.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"ip_address_in": schema.ListAttribute{
						Description: "IP list IDs that the request IP must match.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"ip_address_not_in": schema.ListAttribute{
						Description: "IP list IDs that the request IP must not match.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"device_trusted": schema.BoolAttribute{
						Description: "Whether the device must be trusted.",
						Computed:    true,
					},
					"geo_ip_in": schema.ListAttribute{
						Description: "Country codes the request must originate from.",
						Computed:    true,
						ElementType: types.StringType,
					},
					"geo_ip_not_in": schema.ListAttribute{
						Description: "Country codes the request must not originate from.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"effect": schema.SingleNestedAttribute{
				Description: "The effect when the policy conditions are met.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"action": schema.StringAttribute{
						Description: "The action to take: allow or deny.",
						Computed:    true,
					},
					"mfa_required": schema.BoolAttribute{
						Description: "Whether MFA is required.",
						Computed:    true,
					},
					"mfa_allowed_factors": schema.ListAttribute{
						Description: "Allowed MFA factor types.",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
			"targets": schema.ListNestedAttribute{
				Description: "Target resources this policy applies to.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The target resource type.",
							Computed:    true,
						},
						"id": schema.StringAttribute{
							Description: "The target resource ID.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *AuthenticationPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AuthenticationPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config AuthenticationPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.GetAuthenticationPolicy(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading authentication policy", err.Error())
		return
	}

	config.Name = types.StringValue(policy.Name)
	config.Type = types.StringValue(policy.Type)
	config.Disabled = types.BoolValue(policy.Disabled)

	if policy.Conditions != nil {
		condAttrs := map[string]attr.Value{}

		if len(policy.Conditions.IPAddressIn) > 0 {
			elems := make([]attr.Value, len(policy.Conditions.IPAddressIn))
			for i, v := range policy.Conditions.IPAddressIn {
				elems[i] = types.StringValue(v)
			}
			condAttrs["ip_address_in"], _ = types.ListValue(types.StringType, elems)
		} else {
			condAttrs["ip_address_in"] = types.ListNull(types.StringType)
		}

		if len(policy.Conditions.IPAddressNotIn) > 0 {
			elems := make([]attr.Value, len(policy.Conditions.IPAddressNotIn))
			for i, v := range policy.Conditions.IPAddressNotIn {
				elems[i] = types.StringValue(v)
			}
			condAttrs["ip_address_not_in"], _ = types.ListValue(types.StringType, elems)
		} else {
			condAttrs["ip_address_not_in"] = types.ListNull(types.StringType)
		}

		if policy.Conditions.DeviceTrusted != nil {
			condAttrs["device_trusted"] = types.BoolValue(*policy.Conditions.DeviceTrusted)
		} else {
			condAttrs["device_trusted"] = types.BoolNull()
		}

		if len(policy.Conditions.GeoIPIn) > 0 {
			elems := make([]attr.Value, len(policy.Conditions.GeoIPIn))
			for i, v := range policy.Conditions.GeoIPIn {
				elems[i] = types.StringValue(v)
			}
			condAttrs["geo_ip_in"], _ = types.ListValue(types.StringType, elems)
		} else {
			condAttrs["geo_ip_in"] = types.ListNull(types.StringType)
		}

		if len(policy.Conditions.GeoIPNotIn) > 0 {
			elems := make([]attr.Value, len(policy.Conditions.GeoIPNotIn))
			for i, v := range policy.Conditions.GeoIPNotIn {
				elems[i] = types.StringValue(v)
			}
			condAttrs["geo_ip_not_in"], _ = types.ListValue(types.StringType, elems)
		} else {
			condAttrs["geo_ip_not_in"] = types.ListNull(types.StringType)
		}

		config.Conditions, _ = types.ObjectValue(dsConditionsAttrTypes(), condAttrs)
	} else {
		config.Conditions = types.ObjectNull(dsConditionsAttrTypes())
	}

	if policy.Effect != nil {
		effectAttrs := map[string]attr.Value{
			"action": types.StringValue(policy.Effect.Action),
		}

		if policy.Effect.Obligations != nil && policy.Effect.Obligations.MFA != nil {
			if policy.Effect.Obligations.MFA.Required != nil {
				effectAttrs["mfa_required"] = types.BoolValue(*policy.Effect.Obligations.MFA.Required)
			} else {
				effectAttrs["mfa_required"] = types.BoolNull()
			}
			if len(policy.Effect.Obligations.MFA.AllowedFactors) > 0 {
				elems := make([]attr.Value, len(policy.Effect.Obligations.MFA.AllowedFactors))
				for i, v := range policy.Effect.Obligations.MFA.AllowedFactors {
					elems[i] = types.StringValue(v)
				}
				effectAttrs["mfa_allowed_factors"], _ = types.ListValue(types.StringType, elems)
			} else {
				effectAttrs["mfa_allowed_factors"] = types.ListNull(types.StringType)
			}
		} else {
			effectAttrs["mfa_required"] = types.BoolNull()
			effectAttrs["mfa_allowed_factors"] = types.ListNull(types.StringType)
		}

		config.Effect, _ = types.ObjectValue(dsEffectAttrTypes(), effectAttrs)
	} else {
		config.Effect = types.ObjectNull(dsEffectAttrTypes())
	}

	if policy.Targets != nil && len(policy.Targets.Resources) > 0 {
		targetObjType := types.ObjectType{AttrTypes: dsTargetAttrTypes()}
		targetValues := make([]attr.Value, len(policy.Targets.Resources))
		for i, t := range policy.Targets.Resources {
			targetValues[i], _ = types.ObjectValue(dsTargetAttrTypes(), map[string]attr.Value{
				"type": types.StringValue(t.Type),
				"id":   types.StringValue(t.ID),
			})
		}
		config.Targets, _ = types.ListValue(targetObjType, targetValues)
	} else {
		config.Targets = types.ListNull(types.ObjectType{AttrTypes: dsTargetAttrTypes()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func dsConditionsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_address_in":     types.ListType{ElemType: types.StringType},
		"ip_address_not_in": types.ListType{ElemType: types.StringType},
		"device_trusted":    types.BoolType,
		"geo_ip_in":         types.ListType{ElemType: types.StringType},
		"geo_ip_not_in":     types.ListType{ElemType: types.StringType},
	}
}

func dsEffectAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"action":              types.StringType,
		"mfa_required":        types.BoolType,
		"mfa_allowed_factors": types.ListType{ElemType: types.StringType},
	}
}

func dsTargetAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"id":   types.StringType,
	}
}
