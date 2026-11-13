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

var _ datasource.DataSource = &ApplicationDataSource{}

type ApplicationDataSource struct {
	client *client.Client
}

type ApplicationDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	DisplayLabel types.String `tfsdk:"display_label"`
	SsoURL       types.String `tfsdk:"sso_url"`
	Logo         types.Object `tfsdk:"logo"`
	SSO          types.Object `tfsdk:"sso"`
}

func NewApplicationDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

func (d *ApplicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the application.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the application.",
				Computed:    true,
			},
			"display_label": schema.StringAttribute{
				Description: "Display label shown in the user portal.",
				Computed:    true,
			},
			"sso_url": schema.StringAttribute{
				Description: "SSO URL for the application.",
				Computed:    true,
			},
			"logo": schema.SingleNestedAttribute{
				Description: "Application logo configuration.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "URL of the logo image.",
						Computed:    true,
					},
				},
			},
			"sso": schema.SingleNestedAttribute{
				Description: "SSO settings.",
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "SSO type: saml or oidc.",
						Computed:    true,
					},
					"sp_entity_id": schema.StringAttribute{
						Description: "Service provider entity ID.",
						Computed:    true,
					},
					"acs_url": schema.StringAttribute{
						Description: "Assertion consumer service URL.",
						Computed:    true,
					},
					"idp_entity_id": schema.StringAttribute{
						Description: "Identity provider entity ID.",
						Computed:    true,
					},
					"client_id": schema.StringAttribute{
						Description: "Client ID (OIDC).",
						Computed:    true,
					},
					"redirect_uris": schema.ListAttribute{
						Description: "Redirect URIs (OIDC).",
						Computed:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

func (d *ApplicationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ApplicationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.client.GetApplication(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading application", err.Error())
		return
	}

	config.Name = types.StringValue(app.Name)
	if app.DisplayLabel != "" {
		config.DisplayLabel = types.StringValue(app.DisplayLabel)
	} else {
		config.DisplayLabel = types.StringNull()
	}
	if app.SsoURL != "" {
		config.SsoURL = types.StringValue(app.SsoURL)
	} else {
		config.SsoURL = types.StringNull()
	}

	if app.Logo != nil {
		logoAttrs := map[string]attr.Value{}
		if app.Logo.URL != "" {
			logoAttrs["url"] = types.StringValue(app.Logo.URL)
		} else {
			logoAttrs["url"] = types.StringNull()
		}
		config.Logo, _ = types.ObjectValue(dsAppLogoAttrTypes(), logoAttrs)
	} else {
		config.Logo = types.ObjectNull(dsAppLogoAttrTypes())
	}

	if app.SSO != nil {
		ssoAttrs := map[string]attr.Value{
			"type": types.StringValue(app.SSO.Type),
		}
		if app.SSO.SpEntityID != "" {
			ssoAttrs["sp_entity_id"] = types.StringValue(app.SSO.SpEntityID)
		} else {
			ssoAttrs["sp_entity_id"] = types.StringNull()
		}
		if app.SSO.AcsURL != "" {
			ssoAttrs["acs_url"] = types.StringValue(app.SSO.AcsURL)
		} else {
			ssoAttrs["acs_url"] = types.StringNull()
		}
		if app.SSO.IdpEntityID != "" {
			ssoAttrs["idp_entity_id"] = types.StringValue(app.SSO.IdpEntityID)
		} else {
			ssoAttrs["idp_entity_id"] = types.StringNull()
		}
		if app.SSO.ClientID != "" {
			ssoAttrs["client_id"] = types.StringValue(app.SSO.ClientID)
		} else {
			ssoAttrs["client_id"] = types.StringNull()
		}
		if len(app.SSO.RedirectURIs) > 0 {
			elems := make([]attr.Value, len(app.SSO.RedirectURIs))
			for i, u := range app.SSO.RedirectURIs {
				elems[i] = types.StringValue(u)
			}
			ssoAttrs["redirect_uris"], _ = types.ListValue(types.StringType, elems)
		} else {
			ssoAttrs["redirect_uris"] = types.ListNull(types.StringType)
		}
		config.SSO, _ = types.ObjectValue(dsAppSSOAttrTypes(), ssoAttrs)
	} else {
		config.SSO = types.ObjectNull(dsAppSSOAttrTypes())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func dsAppLogoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"url": types.StringType,
	}
}

func dsAppSSOAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":          types.StringType,
		"sp_entity_id":  types.StringType,
		"acs_url":       types.StringType,
		"idp_entity_id": types.StringType,
		"client_id":     types.StringType,
		"redirect_uris": types.ListType{ElemType: types.StringType},
	}
}
