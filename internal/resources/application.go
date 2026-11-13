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

var _ resource.Resource = &ApplicationResource{}
var _ resource.ResourceWithImportState = &ApplicationResource{}

type ApplicationResource struct {
	client *client.Client
}

type ApplicationResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	DisplayLabel types.String `tfsdk:"display_label"`
	SsoURL       types.String `tfsdk:"sso_url"`
	Logo         types.Object `tfsdk:"logo"`
	SSO          types.Object `tfsdk:"sso"`
	Provision    types.Object `tfsdk:"provision"`
}

func NewApplicationResource() resource.Resource {
	return &ApplicationResource{}
}

func (r *ApplicationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a JumpCloud SSO application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the application.",
				Required:    true,
			},
			"display_label": schema.StringAttribute{
				Description: "Display label shown in the user portal.",
				Optional:    true,
			},
			"sso_url": schema.StringAttribute{
				Description: "SSO URL for the application.",
				Optional:    true,
			},
			"logo": schema.SingleNestedAttribute{
				Description: "Application logo configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"url": schema.StringAttribute{
						Description: "URL of the logo image.",
						Optional:    true,
					},
				},
			},
			"sso": schema.SingleNestedAttribute{
				Description: "SSO settings (SAML/OIDC configuration).",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "SSO type: saml or oidc.",
						Required:    true,
					},
					"sp_entity_id": schema.StringAttribute{
						Description: "Service provider entity ID (SAML).",
						Optional:    true,
					},
					"acs_url": schema.StringAttribute{
						Description: "Assertion consumer service URL (SAML).",
						Optional:    true,
					},
					"idp_entity_id": schema.StringAttribute{
						Description: "Identity provider entity ID (SAML, computed).",
						Optional:    true,
						Computed:    true,
					},
					"idp_cert": schema.StringAttribute{
						Description: "Identity provider certificate (SAML, computed).",
						Optional:    true,
						Computed:    true,
						Sensitive:   true,
					},
					"client_id": schema.StringAttribute{
						Description: "Client ID (OIDC, computed).",
						Optional:    true,
						Computed:    true,
					},
					"client_secret": schema.StringAttribute{
						Description: "Client secret (OIDC, computed).",
						Optional:    true,
						Computed:    true,
						Sensitive:   true,
					},
					"redirect_uris": schema.ListAttribute{
						Description: "Redirect URIs (OIDC).",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
			"provision": schema.SingleNestedAttribute{
				Description: "Provisioning/SCIM settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Description: "Whether provisioning is enabled.",
						Required:    true,
					},
					"type": schema.StringAttribute{
						Description: "Provisioning type (e.g. scim).",
						Optional:    true,
					},
					"base_url": schema.StringAttribute{
						Description: "SCIM base URL.",
						Optional:    true,
					},
					"token": schema.StringAttribute{
						Description: "SCIM bearer token.",
						Optional:    true,
						Sensitive:   true,
					},
				},
			},
		},
	}
}

func (r *ApplicationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app := r.modelToAPI(ctx, &plan)
	created, err := r.client.CreateApplication(ctx, app)
	if err != nil {
		resp.Diagnostics.AddError("Error creating application", err.Error())
		return
	}

	r.apiToModel(created, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetApplication(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading application", err.Error())
		return
	}

	r.apiToModel(app, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app := r.modelToAPI(ctx, &plan)
	updated, err := r.client.UpdateApplication(ctx, state.ID.ValueString(), app)
	if err != nil {
		resp.Diagnostics.AddError("Error updating application", err.Error())
		return
	}

	r.apiToModel(updated, &plan)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApplication(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting application", err.Error())
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	app, err := r.client.GetApplication(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing application", fmt.Sprintf("Cannot read application %s: %s", req.ID, err))
		return
	}

	var state ApplicationResourceModel
	r.apiToModel(app, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ApplicationResource) modelToAPI(ctx context.Context, model *ApplicationResourceModel) *client.Application {
	app := &client.Application{
		Name: model.Name.ValueString(),
	}

	if !model.DisplayLabel.IsNull() && !model.DisplayLabel.IsUnknown() {
		app.DisplayLabel = model.DisplayLabel.ValueString()
	}
	if !model.SsoURL.IsNull() && !model.SsoURL.IsUnknown() {
		app.SsoURL = model.SsoURL.ValueString()
	}

	if !model.Logo.IsNull() && !model.Logo.IsUnknown() {
		attrs := model.Logo.Attributes()
		logo := &client.ApplicationLogo{}
		if v, ok := attrString(attrs, "url"); ok {
			logo.URL = v
		}
		app.Logo = logo
	}

	if !model.SSO.IsNull() && !model.SSO.IsUnknown() {
		attrs := model.SSO.Attributes()
		sso := &client.ApplicationSSO{}
		if v, ok := attrString(attrs, "type"); ok {
			sso.Type = v
		}
		if v, ok := attrString(attrs, "sp_entity_id"); ok {
			sso.SpEntityID = v
		}
		if v, ok := attrString(attrs, "acs_url"); ok {
			sso.AcsURL = v
		}
		if v, ok := attrString(attrs, "idp_entity_id"); ok {
			sso.IdpEntityID = v
		}
		if v, ok := attrString(attrs, "idp_cert"); ok {
			sso.IdpCert = v
		}
		if v, ok := attrString(attrs, "client_id"); ok {
			sso.ClientID = v
		}
		if v, ok := attrString(attrs, "client_secret"); ok {
			sso.ClientSecret = v
		}
		if v, ok := attrStringList(ctx, attrs, "redirect_uris"); ok {
			sso.RedirectURIs = v
		}
		app.SSO = sso
	}

	if !model.Provision.IsNull() && !model.Provision.IsUnknown() {
		attrs := model.Provision.Attributes()
		prov := &client.ApplicationProvisioning{}
		if v, ok := attrBool(attrs, "enabled"); ok {
			prov.Enabled = v
		}
		if v, ok := attrString(attrs, "type"); ok {
			prov.Type = v
		}
		if v, ok := attrString(attrs, "base_url"); ok {
			prov.BaseURL = v
		}
		if v, ok := attrString(attrs, "token"); ok {
			prov.Token = v
		}
		app.Provision = prov
	}

	return app
}

func (r *ApplicationResource) apiToModel(app *client.Application, model *ApplicationResourceModel) {
	model.ID = types.StringValue(app.ID)
	model.Name = types.StringValue(app.Name)

	if app.DisplayLabel != "" {
		model.DisplayLabel = types.StringValue(app.DisplayLabel)
	} else {
		model.DisplayLabel = types.StringNull()
	}
	if app.SsoURL != "" {
		model.SsoURL = types.StringValue(app.SsoURL)
	} else {
		model.SsoURL = types.StringNull()
	}

	if app.Logo != nil {
		logoAttrs := map[string]attr.Value{}
		if app.Logo.URL != "" {
			logoAttrs["url"] = types.StringValue(app.Logo.URL)
		} else {
			logoAttrs["url"] = types.StringNull()
		}
		model.Logo, _ = types.ObjectValue(appLogoAttrTypes(), logoAttrs)
	} else {
		model.Logo = types.ObjectNull(appLogoAttrTypes())
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
		if app.SSO.IdpCert != "" {
			ssoAttrs["idp_cert"] = types.StringValue(app.SSO.IdpCert)
		} else {
			ssoAttrs["idp_cert"] = types.StringNull()
		}
		if app.SSO.ClientID != "" {
			ssoAttrs["client_id"] = types.StringValue(app.SSO.ClientID)
		} else {
			ssoAttrs["client_id"] = types.StringNull()
		}
		if app.SSO.ClientSecret != "" {
			ssoAttrs["client_secret"] = types.StringValue(app.SSO.ClientSecret)
		} else {
			ssoAttrs["client_secret"] = types.StringNull()
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
		model.SSO, _ = types.ObjectValue(appSSOAttrTypes(), ssoAttrs)
	} else {
		model.SSO = types.ObjectNull(appSSOAttrTypes())
	}

	if app.Provision != nil {
		provAttrs := map[string]attr.Value{
			"enabled": types.BoolValue(app.Provision.Enabled),
		}
		if app.Provision.Type != "" {
			provAttrs["type"] = types.StringValue(app.Provision.Type)
		} else {
			provAttrs["type"] = types.StringNull()
		}
		if app.Provision.BaseURL != "" {
			provAttrs["base_url"] = types.StringValue(app.Provision.BaseURL)
		} else {
			provAttrs["base_url"] = types.StringNull()
		}
		if app.Provision.Token != "" {
			provAttrs["token"] = types.StringValue(app.Provision.Token)
		} else {
			provAttrs["token"] = types.StringNull()
		}
		model.Provision, _ = types.ObjectValue(appProvisionAttrTypes(), provAttrs)
	} else {
		model.Provision = types.ObjectNull(appProvisionAttrTypes())
	}
}

func appLogoAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"url": types.StringType,
	}
}

func appSSOAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":          types.StringType,
		"sp_entity_id":  types.StringType,
		"acs_url":       types.StringType,
		"idp_entity_id": types.StringType,
		"idp_cert":      types.StringType,
		"client_id":     types.StringType,
		"client_secret": types.StringType,
		"redirect_uris": types.ListType{ElemType: types.StringType},
	}
}

func appProvisionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled":  types.BoolType,
		"type":     types.StringType,
		"base_url": types.StringType,
		"token":    types.StringType,
	}
}
