// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/datasources"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/resources"
)

// Ensure JumpCloudProvider satisfies various provider interfaces.
var _ provider.Provider = &JumpCloudProvider{}

// JumpCloudProvider defines the provider implementation.
type JumpCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// JumpCloudProviderModel describes the provider data model.
type JumpCloudProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	OrgID   types.String `tfsdk:"org_id"`
	BaseURL types.String `tfsdk:"base_url"`
}

func (p *JumpCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "jumpcloud"
	resp.Version = p.version
}

func (p *JumpCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Interact with JumpCloud APIs to manage user groups, approval flows, commands, policies, and more.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "JumpCloud API key. Can also be set via the `JUMPCLOUD_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"org_id": schema.StringAttribute{
				MarkdownDescription: "JumpCloud organization ID. Required for multi-tenant API keys. Can also be set via the `JUMPCLOUD_ORG_ID` environment variable.",
				Optional:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "JumpCloud API base URL. Defaults to `https://console.jumpcloud.com/api/v2`. Can also be set via the `JUMPCLOUD_BASE_URL` environment variable.",
				Optional:            true,
			},
		},
	}
}

// Configure resolves provider settings in order: explicit configuration,
// then environment variable, then default (or error for the required API key).
func (p *JumpCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data JumpCloudProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("JUMPCLOUD_API_KEY")
	if !data.APIKey.IsNull() {
		apiKey = data.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider cannot create the JumpCloud client because the API key is missing. "+
				"Set the api_key attribute in the provider block or the JUMPCLOUD_API_KEY environment variable.",
		)
		return
	}

	orgID := os.Getenv("JUMPCLOUD_ORG_ID")
	if !data.OrgID.IsNull() {
		orgID = data.OrgID.ValueString()
	}

	baseURL := os.Getenv("JUMPCLOUD_BASE_URL")
	if !data.BaseURL.IsNull() {
		baseURL = data.BaseURL.ValueString()
	}

	c := client.New(apiKey, orgID, baseURL)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *JumpCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewUserGroupResource,
		resources.NewUserGroupMembershipResource,
		resources.NewApprovalFlowResource,
		resources.NewApprovalFlowSettingsResource,
		resources.NewCommandResource,
		resources.NewCommandAssociationResource,
		resources.NewPolicyResource,
		resources.NewPolicyAssociationResource,
		resources.NewPasswordPolicyResource,
		resources.NewPolicyGroupResource,
		resources.NewPolicyGroupMembershipResource,
		resources.NewPolicyGroupAssociationResource,
		resources.NewSoftwareAppResource,
		resources.NewSoftwareAppAssociationResource,
		resources.NewScheduledReportResource,
		resources.NewNotificationChannelResource,
		resources.NewSystemGroupResource,
		resources.NewSystemGroupMembershipResource,
		resources.NewIPListResource,
		resources.NewAuthenticationPolicyResource,
		resources.NewRoleResource,
		resources.NewApplicationResource,
		resources.NewApplicationAssociationResource,
	}
}

func (p *JumpCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewUserGroupDataSource,
		datasources.NewApprovalFlowDataSource,
		datasources.NewCommandDataSource,
		datasources.NewPolicyDataSource,
		datasources.NewPolicyTemplateDataSource,
		datasources.NewPasswordPolicyDataSource,
		datasources.NewPolicyGroupDataSource,
		datasources.NewSoftwareAppDataSource,
		datasources.NewScheduledReportDataSource,
		datasources.NewNotificationChannelDataSource,
		datasources.NewSystemGroupDataSource,
		datasources.NewIPListDataSource,
		datasources.NewAuthenticationPolicyDataSource,
		datasources.NewRoleDataSource,
		datasources.NewApplicationDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &JumpCloudProvider{
			version: version,
		}
	}
}
