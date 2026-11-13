// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

var _ datasource.DataSource = &SoftwareAppDataSource{}

type SoftwareAppDataSource struct {
	client *client.Client
}

type SoftwareAppDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Settings    types.String `tfsdk:"settings"`
}

func NewSoftwareAppDataSource() datasource.DataSource {
	return &SoftwareAppDataSource{}
}

func (d *SoftwareAppDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_software_app"
}

func (d *SoftwareAppDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read a JumpCloud software application by ID or display name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the software application. Exactly one of id or display_name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"display_name": schema.StringAttribute{
				Description: "The display name of the software application. Exactly one of id or display_name must be specified.",
				Optional:    true,
				Computed:    true,
			},
			"settings": schema.StringAttribute{
				Description: "JSON-encoded settings for the software application.",
				Computed:    true,
			},
		},
	}
}

func (d *SoftwareAppDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SoftwareAppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SoftwareAppDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasID := !config.ID.IsNull() && !config.ID.IsUnknown()
	hasName := !config.DisplayName.IsNull() && !config.DisplayName.IsUnknown()

	if !hasID && !hasName {
		resp.Diagnostics.AddError("Missing attribute", "Exactly one of 'id' or 'display_name' must be specified.")
		return
	}
	if hasID && hasName {
		resp.Diagnostics.AddError("Conflicting attributes", "Only one of 'id' or 'display_name' may be specified.")
		return
	}

	var app *client.SoftwareApp
	var err error

	if hasID {
		app, err = d.client.GetSoftwareApp(ctx, config.ID.ValueString())
	} else {
		app, err = d.client.GetSoftwareAppByName(ctx, config.DisplayName.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError("Error reading software app", err.Error())
		return
	}

	config.ID = types.StringValue(app.ID)
	config.DisplayName = types.StringValue(app.DisplayName)
	if len(app.Settings) > 0 && string(app.Settings) != "null" {
		var normalized any
		if err := json.Unmarshal(app.Settings, &normalized); err == nil {
			compact, _ := json.Marshal(normalized)
			config.Settings = types.StringValue(string(compact))
		} else {
			config.Settings = types.StringValue(string(app.Settings))
		}
	} else {
		config.Settings = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
