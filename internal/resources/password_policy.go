// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ resource.Resource = &PasswordPolicyResource{}
var _ resource.ResourceWithImportState = &PasswordPolicyResource{}

type PasswordPolicyResource struct {
	client *client.Client
}

type PasswordPolicyResourceModel struct {
	ID                                     types.String `tfsdk:"id"`
	Name                                   types.String `tfsdk:"name"`
	Description                            types.String `tfsdk:"description"`
	EnableMinLength                        types.Bool   `tfsdk:"enable_min_length"`
	MinLength                              types.Int64  `tfsdk:"min_length"`
	NeedsUppercase                         types.Bool   `tfsdk:"needs_uppercase"`
	NeedsLowercase                         types.Bool   `tfsdk:"needs_lowercase"`
	NeedsNumeric                           types.Bool   `tfsdk:"needs_numeric"`
	NeedsSymbolic                          types.Bool   `tfsdk:"needs_symbolic"`
	EnableMaxLoginAttempts                 types.Bool   `tfsdk:"enable_max_login_attempts"`
	MaxLoginAttempts                       types.Int64  `tfsdk:"max_login_attempts"`
	EnableLockoutTimeInSeconds             types.Bool   `tfsdk:"enable_lockout_time_in_seconds"`
	LockoutTimeInSeconds                   types.Int64  `tfsdk:"lockout_time_in_seconds"`
	EnablePasswordExpirationInDays         types.Bool   `tfsdk:"enable_password_expiration_in_days"`
	PasswordExpirationInDays               types.Int64  `tfsdk:"password_expiration_in_days"`
	EnableDaysBeforeExpirationToForceReset types.Bool   `tfsdk:"enable_days_before_expiration_to_force_reset"`
	DaysBeforeExpirationToForceReset       types.Int64  `tfsdk:"days_before_expiration_to_force_reset"`
	EnableDaysAfterExpirationToSelfRecover types.Bool   `tfsdk:"enable_days_after_expiration_to_self_recover"`
	DaysAfterExpirationToSelfRecover       types.Int64  `tfsdk:"days_after_expiration_to_self_recover"`
	EnableMinChangePeriodInDays            types.Bool   `tfsdk:"enable_min_change_period_in_days"`
	MinChangePeriodInDays                  types.Int64  `tfsdk:"min_change_period_in_days"`
	EnableMaxHistory                       types.Bool   `tfsdk:"enable_max_history"`
	MaxHistory                             types.Int64  `tfsdk:"max_history"`
	EnableResetLockoutCounter              types.Bool   `tfsdk:"enable_reset_lockout_counter"`
	ResetLockoutCounterMinutes             types.Int64  `tfsdk:"reset_lockout_counter_minutes"`
	DisallowCommonlyUsedPasswords          types.Bool   `tfsdk:"disallow_commonly_used_passwords"`
	DisallowSequentialOrRepetitiveChars    types.Bool   `tfsdk:"disallow_sequential_or_repetitive_chars"`
	AllowUsernameSubstring                 types.Bool   `tfsdk:"allow_username_substring"`
	AllowUnenrolledMFAPasswordReset        types.Bool   `tfsdk:"allow_unenrolled_mfa_password_reset"`
	EnableRecoveryEmail                    types.Bool   `tfsdk:"enable_recovery_email"`
	DisplayComplexityOnResetScreen         types.Bool   `tfsdk:"display_complexity_on_reset_screen"`
	Precedence                             types.Int64  `tfsdk:"precedence"`
}

func NewPasswordPolicyResource() resource.Resource {
	return &PasswordPolicyResource{}
}

func (r *PasswordPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password_policy"
}

func (r *PasswordPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud password policy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the password policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the password policy.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the password policy.",
				Optional:    true,
			},
			"enable_min_length": schema.BoolAttribute{
				Description: "Whether minimum length enforcement is enabled.",
				Optional:    true,
			},
			"min_length": schema.Int64Attribute{
				Description: "Minimum password length.",
				Optional:    true,
			},
			"needs_uppercase": schema.BoolAttribute{
				Description: "Whether the password must contain uppercase characters.",
				Optional:    true,
			},
			"needs_lowercase": schema.BoolAttribute{
				Description: "Whether the password must contain lowercase characters.",
				Optional:    true,
			},
			"needs_numeric": schema.BoolAttribute{
				Description: "Whether the password must contain numeric characters.",
				Optional:    true,
			},
			"needs_symbolic": schema.BoolAttribute{
				Description: "Whether the password must contain special characters.",
				Optional:    true,
			},
			"enable_max_login_attempts": schema.BoolAttribute{
				Description: "Whether login attempt limiting is enabled.",
				Optional:    true,
			},
			"max_login_attempts": schema.Int64Attribute{
				Description: "Number of failed attempts before account lockout.",
				Optional:    true,
			},
			"enable_lockout_time_in_seconds": schema.BoolAttribute{
				Description: "Whether timed lockout is enabled.",
				Optional:    true,
			},
			"lockout_time_in_seconds": schema.Int64Attribute{
				Description: "Duration of account lockout in seconds.",
				Optional:    true,
			},
			"enable_password_expiration_in_days": schema.BoolAttribute{
				Description: "Whether password expiration is enabled.",
				Optional:    true,
			},
			"password_expiration_in_days": schema.Int64Attribute{
				Description: "Number of days until the password expires.",
				Optional:    true,
			},
			"enable_days_before_expiration_to_force_reset": schema.BoolAttribute{
				Description: "Whether forced reset before expiration is enabled.",
				Optional:    true,
			},
			"days_before_expiration_to_force_reset": schema.Int64Attribute{
				Description: "Days before expiration to force a password reset.",
				Optional:    true,
			},
			"enable_days_after_expiration_to_self_recover": schema.BoolAttribute{
				Description: "Whether self-recovery after expiration is enabled.",
				Optional:    true,
			},
			"days_after_expiration_to_self_recover": schema.Int64Attribute{
				Description: "Days after expiration a user can self-recover. -1 means unlimited.",
				Optional:    true,
			},
			"enable_min_change_period_in_days": schema.BoolAttribute{
				Description: "Whether minimum change period enforcement is enabled.",
				Optional:    true,
			},
			"min_change_period_in_days": schema.Int64Attribute{
				Description: "Minimum days between password changes.",
				Optional:    true,
			},
			"enable_max_history": schema.BoolAttribute{
				Description: "Whether password history enforcement is enabled.",
				Optional:    true,
			},
			"max_history": schema.Int64Attribute{
				Description: "Number of previous passwords to remember.",
				Optional:    true,
			},
			"enable_reset_lockout_counter": schema.BoolAttribute{
				Description: "Whether lockout counter reset is enabled.",
				Optional:    true,
			},
			"reset_lockout_counter_minutes": schema.Int64Attribute{
				Description: "Minutes after which the lockout counter resets.",
				Optional:    true,
			},
			"disallow_commonly_used_passwords": schema.BoolAttribute{
				Description: "Whether commonly used passwords are disallowed.",
				Optional:    true,
			},
			"disallow_sequential_or_repetitive_chars": schema.BoolAttribute{
				Description: "Whether sequential or repetitive characters are disallowed.",
				Optional:    true,
			},
			"allow_username_substring": schema.BoolAttribute{
				Description: "Whether the password can contain the username.",
				Optional:    true,
			},
			"allow_unenrolled_mfa_password_reset": schema.BoolAttribute{
				Description: "Whether users not enrolled in MFA can reset their password.",
				Optional:    true,
			},
			"enable_recovery_email": schema.BoolAttribute{
				Description: "Whether recovery email is enabled.",
				Optional:    true,
			},
			"display_complexity_on_reset_screen": schema.BoolAttribute{
				Description: "Whether to display password complexity requirements on the reset screen.",
				Optional:    true,
			},
			"precedence": schema.Int64Attribute{
				Description: "Priority when multiple policies apply. Higher values take precedence.",
				Optional:    true,
			},
		},
	}
}

func (r *PasswordPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PasswordPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PasswordPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := r.modelToAPI(&plan)
	created, err := r.client.CreatePasswordPolicy(ctx, policy)
	if err != nil {
		resp.Diagnostics.AddError("Error creating password policy", err.Error())
		return
	}

	plan.ID = types.StringValue(created.ID)
	r.apiToModel(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PasswordPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PasswordPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetPasswordPolicy(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading password policy", err.Error())
		return
	}

	r.apiToModel(policy, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PasswordPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PasswordPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state PasswordPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy := r.modelToAPI(&plan)
	updated, err := r.client.UpdatePasswordPolicy(ctx, state.ID.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError("Error updating password policy", err.Error())
		return
	}

	plan.ID = state.ID
	r.apiToModel(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PasswordPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PasswordPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePasswordPolicy(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting password policy", err.Error())
		return
	}
}

func (r *PasswordPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	policy, err := r.client.GetPasswordPolicy(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing password policy", fmt.Sprintf("Cannot read password policy %s: %s", req.ID, err))
		return
	}

	var state PasswordPolicyResourceModel
	state.ID = types.StringValue(policy.ID)
	r.apiToModel(policy, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PasswordPolicyResource) modelToAPI(model *PasswordPolicyResourceModel) *client.PasswordPolicy {
	policy := &client.PasswordPolicy{
		Name: model.Name.ValueString(),
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		policy.Description = model.Description.ValueString()
	}
	if !model.EnableMinLength.IsNull() {
		policy.EnableMinLength = model.EnableMinLength.ValueBool()
	}
	if !model.MinLength.IsNull() && !model.MinLength.IsUnknown() {
		policy.MinLength = int(model.MinLength.ValueInt64())
	}
	if !model.NeedsUppercase.IsNull() {
		policy.NeedsUppercase = model.NeedsUppercase.ValueBool()
	}
	if !model.NeedsLowercase.IsNull() {
		policy.NeedsLowercase = model.NeedsLowercase.ValueBool()
	}
	if !model.NeedsNumeric.IsNull() {
		policy.NeedsNumeric = model.NeedsNumeric.ValueBool()
	}
	if !model.NeedsSymbolic.IsNull() {
		policy.NeedsSymbolic = model.NeedsSymbolic.ValueBool()
	}
	if !model.EnableMaxLoginAttempts.IsNull() {
		policy.EnableMaxLoginAttempts = model.EnableMaxLoginAttempts.ValueBool()
	}
	if !model.MaxLoginAttempts.IsNull() && !model.MaxLoginAttempts.IsUnknown() {
		policy.MaxLoginAttempts = int(model.MaxLoginAttempts.ValueInt64())
	}
	if !model.EnableLockoutTimeInSeconds.IsNull() {
		policy.EnableLockoutTimeInSeconds = model.EnableLockoutTimeInSeconds.ValueBool()
	}
	if !model.LockoutTimeInSeconds.IsNull() && !model.LockoutTimeInSeconds.IsUnknown() {
		policy.LockoutTimeInSeconds = int(model.LockoutTimeInSeconds.ValueInt64())
	}
	if !model.EnablePasswordExpirationInDays.IsNull() {
		policy.EnablePasswordExpirationInDays = model.EnablePasswordExpirationInDays.ValueBool()
	}
	if !model.PasswordExpirationInDays.IsNull() && !model.PasswordExpirationInDays.IsUnknown() {
		policy.PasswordExpirationInDays = int(model.PasswordExpirationInDays.ValueInt64())
	}
	if !model.EnableDaysBeforeExpirationToForceReset.IsNull() {
		policy.EnableDaysBeforeExpirationToForceReset = model.EnableDaysBeforeExpirationToForceReset.ValueBool()
	}
	if !model.DaysBeforeExpirationToForceReset.IsNull() && !model.DaysBeforeExpirationToForceReset.IsUnknown() {
		policy.DaysBeforeExpirationToForceReset = int(model.DaysBeforeExpirationToForceReset.ValueInt64())
	}
	if !model.EnableDaysAfterExpirationToSelfRecover.IsNull() {
		policy.EnableDaysAfterExpirationToSelfRecover = model.EnableDaysAfterExpirationToSelfRecover.ValueBool()
	}
	if !model.DaysAfterExpirationToSelfRecover.IsNull() && !model.DaysAfterExpirationToSelfRecover.IsUnknown() {
		policy.DaysAfterExpirationToSelfRecover = int(model.DaysAfterExpirationToSelfRecover.ValueInt64())
	}
	if !model.EnableMinChangePeriodInDays.IsNull() {
		policy.EnableMinChangePeriodInDays = model.EnableMinChangePeriodInDays.ValueBool()
	}
	if !model.MinChangePeriodInDays.IsNull() && !model.MinChangePeriodInDays.IsUnknown() {
		policy.MinChangePeriodInDays = int(model.MinChangePeriodInDays.ValueInt64())
	}
	if !model.EnableMaxHistory.IsNull() {
		policy.EnableMaxHistory = model.EnableMaxHistory.ValueBool()
	}
	if !model.MaxHistory.IsNull() && !model.MaxHistory.IsUnknown() {
		policy.MaxHistory = int(model.MaxHistory.ValueInt64())
	}
	if !model.EnableResetLockoutCounter.IsNull() {
		policy.EnableResetLockoutCounter = model.EnableResetLockoutCounter.ValueBool()
	}
	if !model.ResetLockoutCounterMinutes.IsNull() && !model.ResetLockoutCounterMinutes.IsUnknown() {
		policy.ResetLockoutCounterMinutes = int(model.ResetLockoutCounterMinutes.ValueInt64())
	}
	if !model.DisallowCommonlyUsedPasswords.IsNull() {
		policy.DisallowCommonlyUsedPasswords = model.DisallowCommonlyUsedPasswords.ValueBool()
	}
	if !model.DisallowSequentialOrRepetitiveChars.IsNull() {
		policy.DisallowSequentialOrRepetitiveChars = model.DisallowSequentialOrRepetitiveChars.ValueBool()
	}
	if !model.AllowUsernameSubstring.IsNull() {
		policy.AllowUsernameSubstring = model.AllowUsernameSubstring.ValueBool()
	}
	if !model.AllowUnenrolledMFAPasswordReset.IsNull() {
		policy.AllowUnenrolledMFAPasswordReset = model.AllowUnenrolledMFAPasswordReset.ValueBool()
	}
	if !model.EnableRecoveryEmail.IsNull() {
		policy.EnableRecoveryEmail = model.EnableRecoveryEmail.ValueBool()
	}
	if !model.DisplayComplexityOnResetScreen.IsNull() {
		policy.DisplayComplexityOnResetScreen = model.DisplayComplexityOnResetScreen.ValueBool()
	}
	if !model.Precedence.IsNull() && !model.Precedence.IsUnknown() {
		policy.Precedence = int(model.Precedence.ValueInt64())
	}
	return policy
}

func (r *PasswordPolicyResource) apiToModel(policy *client.PasswordPolicy, model *PasswordPolicyResourceModel) {
	if policy.Name != "" {
		model.Name = types.StringValue(policy.Name)
	} else {
		model.Name = types.StringValue("")
	}
	if policy.Description != "" {
		model.Description = types.StringValue(policy.Description)
	} else {
		model.Description = types.StringNull()
	}

	model.EnableMinLength = types.BoolValue(policy.EnableMinLength)
	if policy.MinLength != 0 {
		model.MinLength = types.Int64Value(int64(policy.MinLength))
	} else {
		model.MinLength = types.Int64Null()
	}

	model.NeedsUppercase = types.BoolValue(policy.NeedsUppercase)
	model.NeedsLowercase = types.BoolValue(policy.NeedsLowercase)
	model.NeedsNumeric = types.BoolValue(policy.NeedsNumeric)
	model.NeedsSymbolic = types.BoolValue(policy.NeedsSymbolic)

	model.EnableMaxLoginAttempts = types.BoolValue(policy.EnableMaxLoginAttempts)
	if policy.MaxLoginAttempts != 0 {
		model.MaxLoginAttempts = types.Int64Value(int64(policy.MaxLoginAttempts))
	} else {
		model.MaxLoginAttempts = types.Int64Null()
	}

	model.EnableLockoutTimeInSeconds = types.BoolValue(policy.EnableLockoutTimeInSeconds)
	if policy.LockoutTimeInSeconds != 0 {
		model.LockoutTimeInSeconds = types.Int64Value(int64(policy.LockoutTimeInSeconds))
	} else {
		model.LockoutTimeInSeconds = types.Int64Null()
	}

	model.EnablePasswordExpirationInDays = types.BoolValue(policy.EnablePasswordExpirationInDays)
	if policy.PasswordExpirationInDays != 0 {
		model.PasswordExpirationInDays = types.Int64Value(int64(policy.PasswordExpirationInDays))
	} else {
		model.PasswordExpirationInDays = types.Int64Null()
	}

	model.EnableDaysBeforeExpirationToForceReset = types.BoolValue(policy.EnableDaysBeforeExpirationToForceReset)
	if policy.DaysBeforeExpirationToForceReset != 0 {
		model.DaysBeforeExpirationToForceReset = types.Int64Value(int64(policy.DaysBeforeExpirationToForceReset))
	} else {
		model.DaysBeforeExpirationToForceReset = types.Int64Null()
	}

	model.EnableDaysAfterExpirationToSelfRecover = types.BoolValue(policy.EnableDaysAfterExpirationToSelfRecover)
	if policy.DaysAfterExpirationToSelfRecover != 0 {
		model.DaysAfterExpirationToSelfRecover = types.Int64Value(int64(policy.DaysAfterExpirationToSelfRecover))
	} else {
		model.DaysAfterExpirationToSelfRecover = types.Int64Null()
	}

	model.EnableMinChangePeriodInDays = types.BoolValue(policy.EnableMinChangePeriodInDays)
	if policy.MinChangePeriodInDays != 0 {
		model.MinChangePeriodInDays = types.Int64Value(int64(policy.MinChangePeriodInDays))
	} else {
		model.MinChangePeriodInDays = types.Int64Null()
	}

	model.EnableMaxHistory = types.BoolValue(policy.EnableMaxHistory)
	if policy.MaxHistory != 0 {
		model.MaxHistory = types.Int64Value(int64(policy.MaxHistory))
	} else {
		model.MaxHistory = types.Int64Null()
	}

	model.EnableResetLockoutCounter = types.BoolValue(policy.EnableResetLockoutCounter)
	if policy.ResetLockoutCounterMinutes != 0 {
		model.ResetLockoutCounterMinutes = types.Int64Value(int64(policy.ResetLockoutCounterMinutes))
	} else {
		model.ResetLockoutCounterMinutes = types.Int64Null()
	}

	model.DisallowCommonlyUsedPasswords = types.BoolValue(policy.DisallowCommonlyUsedPasswords)
	model.DisallowSequentialOrRepetitiveChars = types.BoolValue(policy.DisallowSequentialOrRepetitiveChars)
	model.AllowUsernameSubstring = types.BoolValue(policy.AllowUsernameSubstring)
	model.AllowUnenrolledMFAPasswordReset = types.BoolValue(policy.AllowUnenrolledMFAPasswordReset)
	model.EnableRecoveryEmail = types.BoolValue(policy.EnableRecoveryEmail)
	model.DisplayComplexityOnResetScreen = types.BoolValue(policy.DisplayComplexityOnResetScreen)

	if policy.Precedence != 0 {
		model.Precedence = types.Int64Value(int64(policy.Precedence))
	} else {
		model.Precedence = types.Int64Null()
	}
}
