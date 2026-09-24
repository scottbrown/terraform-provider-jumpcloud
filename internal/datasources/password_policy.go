// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ datasource.DataSource = &PasswordPolicyDataSource{}

type PasswordPolicyDataSource struct {
	client *client.Client
}

type PasswordPolicyDataSourceModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	EnableMinLength          types.Bool   `tfsdk:"enable_min_length"`
	MinLength                types.Int64  `tfsdk:"min_length"`
	NeedsUppercase           types.Bool   `tfsdk:"needs_uppercase"`
	NeedsLowercase           types.Bool   `tfsdk:"needs_lowercase"`
	NeedsNumeric             types.Bool   `tfsdk:"needs_numeric"`
	NeedsSymbolic            types.Bool   `tfsdk:"needs_symbolic"`
	EnableMaxLoginAttempts   types.Bool   `tfsdk:"enable_max_login_attempts"`
	MaxLoginAttempts         types.Int64  `tfsdk:"max_login_attempts"`
	LockoutTimeInSeconds     types.Int64  `tfsdk:"lockout_time_in_seconds"`
	PasswordExpirationInDays types.Int64  `tfsdk:"password_expiration_in_days"`
	Precedence               types.Int64  `tfsdk:"precedence"`
}

func NewPasswordPolicyDataSource() datasource.DataSource {
	return &PasswordPolicyDataSource{}
}

func (d *PasswordPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_password_policy"
}

func (d *PasswordPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud password policy by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the password policy. Exactly one of id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the password policy. Exactly one of id or name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"enable_min_length": schema.BoolAttribute{
				Description: "Whether minimum length enforcement is enabled.",
				Computed:    true,
			},
			"min_length": schema.Int64Attribute{
				Description: "Minimum password length.",
				Computed:    true,
			},
			"needs_uppercase": schema.BoolAttribute{
				Description: "Whether the password must contain uppercase characters.",
				Computed:    true,
			},
			"needs_lowercase": schema.BoolAttribute{
				Description: "Whether the password must contain lowercase characters.",
				Computed:    true,
			},
			"needs_numeric": schema.BoolAttribute{
				Description: "Whether the password must contain numeric characters.",
				Computed:    true,
			},
			"needs_symbolic": schema.BoolAttribute{
				Description: "Whether the password must contain special characters.",
				Computed:    true,
			},
			"enable_max_login_attempts": schema.BoolAttribute{
				Description: "Whether login attempt limiting is enabled.",
				Computed:    true,
			},
			"max_login_attempts": schema.Int64Attribute{
				Description: "Number of failed attempts before account lockout.",
				Computed:    true,
			},
			"lockout_time_in_seconds": schema.Int64Attribute{
				Description: "Duration of account lockout in seconds.",
				Computed:    true,
			},
			"password_expiration_in_days": schema.Int64Attribute{
				Description: "Number of days until the password expires.",
				Computed:    true,
			},
			"precedence": schema.Int64Attribute{
				Description: "Priority when multiple policies apply.",
				Computed:    true,
			},
		},
	}
}

func (d *PasswordPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PasswordPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PasswordPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && !config.ID.IsUnknown()
	hasName := !config.Name.IsNull() && !config.Name.IsUnknown()

	if !hasID && !hasName {
		resp.Diagnostics.AddError("Missing attribute", "Exactly one of 'id' or 'name' must be specified.")
		return
	}
	if hasID && hasName {
		resp.Diagnostics.AddError("Conflicting attributes", "Only one of 'id' or 'name' may be specified.")
		return
	}

	var policy *client.PasswordPolicy
	var err error

	if hasID {
		policy, err = d.client.GetPasswordPolicy(ctx, config.ID.ValueString())
	} else {
		policy, err = d.client.GetPasswordPolicyByName(ctx, config.Name.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Error reading password policy", err.Error())
		return
	}

	config.ID = types.StringValue(policy.ID)
	config.Name = types.StringValue(policy.Name)
	config.EnableMinLength = types.BoolValue(policy.EnableMinLength)
	config.NeedsUppercase = types.BoolValue(policy.NeedsUppercase)
	config.NeedsLowercase = types.BoolValue(policy.NeedsLowercase)
	config.NeedsNumeric = types.BoolValue(policy.NeedsNumeric)
	config.NeedsSymbolic = types.BoolValue(policy.NeedsSymbolic)
	config.EnableMaxLoginAttempts = types.BoolValue(policy.EnableMaxLoginAttempts)

	if policy.MinLength != 0 {
		config.MinLength = types.Int64Value(int64(policy.MinLength))
	}
	if policy.MaxLoginAttempts != 0 {
		config.MaxLoginAttempts = types.Int64Value(int64(policy.MaxLoginAttempts))
	}
	if policy.LockoutTimeInSeconds != 0 {
		config.LockoutTimeInSeconds = types.Int64Value(int64(policy.LockoutTimeInSeconds))
	}
	if policy.PasswordExpirationInDays != 0 {
		config.PasswordExpirationInDays = types.Int64Value(int64(policy.PasswordExpirationInDays))
	}
	if policy.Precedence != 0 {
		config.Precedence = types.Int64Value(int64(policy.Precedence))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
