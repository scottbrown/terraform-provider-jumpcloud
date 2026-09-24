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

var _ resource.Resource = &AuthenticationPolicyResource{}
var _ resource.ResourceWithImportState = &AuthenticationPolicyResource{}

type AuthenticationPolicyResource struct {
	client *client.Client
}

type AuthenticationPolicyResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Disabled   types.Bool   `tfsdk:"disabled"`
	Conditions types.Object `tfsdk:"conditions"`
	Effect     types.Object `tfsdk:"effect"`
	Targets    types.List   `tfsdk:"targets"`
}

func NewAuthenticationPolicyResource() resource.Resource {
	return &AuthenticationPolicyResource{}
}

func (r *AuthenticationPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authentication_policy"
}

func (r *AuthenticationPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud authentication policy for conditional access.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the authentication policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the authentication policy.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "The policy type (e.g. user_portal, resource).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"disabled": schema.BoolAttribute{
				Description: "Whether the policy is disabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"conditions": schema.SingleNestedAttribute{
				Description: "Conditions that determine when this policy applies.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"ip_address_in": schema.ListAttribute{
						Description: "IP list IDs that the request IP must match.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"ip_address_not_in": schema.ListAttribute{
						Description: "IP list IDs that the request IP must not match.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"device_trusted": schema.BoolAttribute{
						Description: "Whether the device must be trusted.",
						Optional:    true,
					},
					"geo_ip_in": schema.ListAttribute{
						Description: "Country codes the request must originate from.",
						Optional:    true,
						ElementType: types.StringType,
					},
					"geo_ip_not_in": schema.ListAttribute{
						Description: "Country codes the request must not originate from.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"effect": schema.SingleNestedAttribute{
				Description: "The effect when the policy conditions are met.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"action": schema.StringAttribute{
						Description: "The action to take: allow or deny.",
						Required:    true,
					},
					"mfa_required": schema.BoolAttribute{
						Description: "Whether MFA is required.",
						Optional:    true,
					},
					"mfa_allowed_factors": schema.ListAttribute{
						Description: "Allowed MFA factor types (e.g. totp, webauthn, push).",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"targets": schema.ListNestedAttribute{
				Description: "Target resources this policy applies to.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Description: "The target resource type.",
							Required:    true,
						},
						"id": schema.StringAttribute{
							Description: "The target resource ID.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *AuthenticationPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AuthenticationPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AuthenticationPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := r.modelToAPI(ctx, &plan)
	created, err := r.client.CreateAuthenticationPolicy(ctx, policy)
	if err != nil {
		resp.Diagnostics.AddError("Error creating authentication policy", err.Error())
		return
	}

	r.apiToModel(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AuthenticationPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AuthenticationPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetAuthenticationPolicy(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading authentication policy", err.Error())
		return
	}

	r.apiToModel(policy, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AuthenticationPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AuthenticationPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AuthenticationPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := r.modelToAPI(ctx, &plan)
	updated, err := r.client.UpdateAuthenticationPolicy(ctx, state.ID.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError("Error updating authentication policy", err.Error())
		return
	}

	r.apiToModel(updated, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AuthenticationPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AuthenticationPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAuthenticationPolicy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting authentication policy", err.Error())
		return
	}
}

func (r *AuthenticationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policy, err := r.client.GetAuthenticationPolicy(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing authentication policy", fmt.Sprintf("Cannot read authentication policy %s: %s", req.ID, err))
		return
	}

	var state AuthenticationPolicyResourceModel
	r.apiToModel(policy, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AuthenticationPolicyResource) modelToAPI(ctx context.Context, model *AuthenticationPolicyResourceModel) *client.AuthenticationPolicy {
	policy := &client.AuthenticationPolicy{
		Name:     model.Name.ValueString(),
		Type:     model.Type.ValueString(),
		Disabled: model.Disabled.ValueBool(),
	}

	if !model.Conditions.IsNull() && !model.Conditions.IsUnknown() {
		conditions := &client.AuthenticationPolicyConditions{}

		attrs := model.Conditions.Attributes()
		if v, ok := attrStringList(ctx, attrs, "ip_address_in"); ok {
			conditions.IPAddressIn = v
		}
		if v, ok := attrStringList(ctx, attrs, "ip_address_not_in"); ok {
			conditions.IPAddressNotIn = v
		}
		if v, ok := attrBool(attrs, "device_trusted"); ok {
			trusted := v
			conditions.DeviceTrusted = &trusted
		}
		if v, ok := attrStringList(ctx, attrs, "geo_ip_in"); ok {
			conditions.GeoIPIn = v
		}
		if v, ok := attrStringList(ctx, attrs, "geo_ip_not_in"); ok {
			conditions.GeoIPNotIn = v
		}

		policy.Conditions = conditions
	}

	if !model.Effect.IsNull() && !model.Effect.IsUnknown() {
		effect := &client.AuthenticationPolicyEffect{}
		attrs := model.Effect.Attributes()

		if v, ok := attrString(attrs, "action"); ok {
			effect.Action = v
		}

		mfaRequired := false
		var mfaFactors []string
		if v, ok := attrBool(attrs, "mfa_required"); ok {
			mfaRequired = v
		}
		if v, ok := attrStringList(ctx, attrs, "mfa_allowed_factors"); ok {
			mfaFactors = v
		}

		if mfaRequired || len(mfaFactors) > 0 {
			effect.Obligations = &client.AuthenticationPolicyEffectObligations{
				MFA: &client.AuthenticationPolicyMFA{},
			}
			if mfaRequired {
				effect.Obligations.MFA.Required = &mfaRequired
			}
			if len(mfaFactors) > 0 {
				effect.Obligations.MFA.AllowedFactors = mfaFactors
			}
		}

		policy.Effect = effect
	}

	if !model.Targets.IsNull() && !model.Targets.IsUnknown() {
		type targetModel struct {
			Type string `tfsdk:"type"`
			ID   string `tfsdk:"id"`
		}
		var targets []targetModel
		model.Targets.ElementsAs(ctx, &targets, false)

		policy.Targets = &client.AuthenticationPolicyTargets{}
		for _, t := range targets {
			policy.Targets.Resources = append(policy.Targets.Resources, client.AuthenticationPolicyTargetResource{
				Type: t.Type,
				ID:   t.ID,
			})
		}
	}

	return policy
}

func (r *AuthenticationPolicyResource) apiToModel(policy *client.AuthenticationPolicy, model *AuthenticationPolicyResourceModel) {
	model.ID = types.StringValue(policy.ID)
	model.Name = types.StringValue(policy.Name)
	model.Type = types.StringValue(policy.Type)
	model.Disabled = types.BoolValue(policy.Disabled)

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

		model.Conditions, _ = types.ObjectValue(conditionsAttrTypes(), condAttrs)
	} else {
		model.Conditions = types.ObjectNull(conditionsAttrTypes())
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

		model.Effect, _ = types.ObjectValue(effectAttrTypes(), effectAttrs)
	} else {
		model.Effect = types.ObjectNull(effectAttrTypes())
	}

	if policy.Targets != nil && len(policy.Targets.Resources) > 0 {
		targetObjType := types.ObjectType{
			AttrTypes: targetAttrTypes(),
		}
		targetValues := make([]attr.Value, len(policy.Targets.Resources))
		for i, t := range policy.Targets.Resources {
			targetValues[i], _ = types.ObjectValue(targetAttrTypes(), map[string]attr.Value{
				"type": types.StringValue(t.Type),
				"id":   types.StringValue(t.ID),
			})
		}
		model.Targets, _ = types.ListValue(targetObjType, targetValues)
	} else {
		model.Targets = types.ListNull(types.ObjectType{AttrTypes: targetAttrTypes()})
	}
}

func conditionsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"ip_address_in":     types.ListType{ElemType: types.StringType},
		"ip_address_not_in": types.ListType{ElemType: types.StringType},
		"device_trusted":    types.BoolType,
		"geo_ip_in":         types.ListType{ElemType: types.StringType},
		"geo_ip_not_in":     types.ListType{ElemType: types.StringType},
	}
}

func effectAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"action":              types.StringType,
		"mfa_required":        types.BoolType,
		"mfa_allowed_factors": types.ListType{ElemType: types.StringType},
	}
}

func targetAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type": types.StringType,
		"id":   types.StringType,
	}
}
