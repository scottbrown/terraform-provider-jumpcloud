// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type Application struct {
	ID           string                   `json:"id,omitempty"`
	Name         string                   `json:"name"`
	DisplayLabel string                   `json:"displayLabel,omitempty"`
	SsoURL       string                   `json:"ssoUrl,omitempty"`
	Logo         *ApplicationLogo         `json:"logo,omitempty"`
	SSO          *ApplicationSSO          `json:"sso,omitempty"`
	Provision    *ApplicationProvisioning `json:"provision,omitempty"`
}

type ApplicationLogo struct {
	URL string `json:"url,omitempty"`
}

type ApplicationSSO struct {
	Type         string   `json:"type,omitempty"`
	SpEntityID   string   `json:"spEntityId,omitempty"`
	AcsURL       string   `json:"acsUrl,omitempty"`
	IdpEntityID  string   `json:"idpEntityId,omitempty"`
	IdpCert      string   `json:"idpCert,omitempty"`
	ClientID     string   `json:"clientId,omitempty"`
	ClientSecret string   `json:"clientSecret,omitempty"`
	RedirectURIs []string `json:"redirectUris,omitempty"`
}

type ApplicationProvisioning struct {
	Enabled bool   `json:"enabled,omitempty"`
	Type    string `json:"type,omitempty"`
	BaseURL string `json:"baseUrl,omitempty"`
	Token   string `json:"token,omitempty"`
}

type ApplicationAssociation struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (a *Application) UnmarshalJSON(data []byte) error {
	// The JumpCloud API returns:
	//   "_id" instead of "id"
	//   "sso.url" for the SSO URL
	//   "sso.oidc.clientId" for OIDC client ID (nested under sso.oidc)
	//   "provision" as an object (presence = enabled) with "_id" and "type"
	var raw struct {
		ID           string `json:"_id"`
		IDalt        string `json:"id"`
		Name         string `json:"name"`
		DisplayLabel string `json:"displayLabel"`
		Logo         *struct {
			URL string `json:"url"`
		} `json:"logo"`
		SSO *struct {
			Type string `json:"type"`
			URL  string `json:"url"`
			OIDC *struct {
				ClientID string `json:"clientId"`
			} `json:"oidc"`
		} `json:"sso"`
		Provision *struct {
			Type string `json:"type"`
		} `json:"provision"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	a.ID = raw.ID
	if a.ID == "" {
		a.ID = raw.IDalt
	}
	a.Name = raw.Name
	a.DisplayLabel = raw.DisplayLabel

	if raw.SSO != nil {
		a.SsoURL = raw.SSO.URL
		a.SSO = &ApplicationSSO{
			Type: raw.SSO.Type,
		}
		if raw.SSO.OIDC != nil {
			a.SSO.ClientID = raw.SSO.OIDC.ClientID
		}
	}

	if raw.Logo != nil && raw.Logo.URL != "" {
		a.Logo = &ApplicationLogo{URL: raw.Logo.URL}
	}

	if raw.Provision != nil {
		a.Provision = &ApplicationProvisioning{
			Enabled: true,
			Type:    raw.Provision.Type,
		}
	}

	return nil
}

func (c *Client) CreateApplication(ctx context.Context, app *Application) (*Application, error) {
	var result Application
	err := c.do(ctx, "POST", "/applications", app, &result)
	if err != nil {
		return nil, fmt.Errorf("creating application: %w", err)
	}
	return &result, nil
}

func (c *Client) GetApplication(ctx context.Context, id string) (*Application, error) {
	var result Application
	err := c.do(ctx, "GET", "/applications/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting application: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateApplication(ctx context.Context, id string, app *Application) (*Application, error) {
	var result Application
	err := c.do(ctx, "PUT", "/applications/"+id, app, &result)
	if err != nil {
		return nil, fmt.Errorf("updating application: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteApplication(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/applications/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting application: %w", err)
	}
	return nil
}

func (c *Client) ListApplications(ctx context.Context, params *PaginationParams) ([]Application, error) {
	path := c.buildPaginatedPath("/applications", params)
	var result []Application
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing applications: %w", err)
	}
	return result, nil
}

func (c *Client) GetApplicationAssociations(ctx context.Context, appID, targetType string) ([]ApplicationAssociation, error) {
	path := fmt.Sprintf("/applications/%s/associations?targets=%s", appID, targetType)
	var result []ApplicationAssociation
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting application associations: %w", err)
	}
	return result, nil
}

func (c *Client) AddApplicationAssociation(ctx context.Context, appID string, assoc *ApplicationAssociation) error {
	path := fmt.Sprintf("/applications/%s/associations", appID)
	body := map[string]any{
		"op":   "add",
		"type": assoc.Type,
		"id":   assoc.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("adding application association: %w", err)
	}
	return nil
}

func (c *Client) RemoveApplicationAssociation(ctx context.Context, appID string, assoc *ApplicationAssociation) error {
	path := fmt.Sprintf("/applications/%s/associations", appID)
	body := map[string]any{
		"op":   "remove",
		"type": assoc.Type,
		"id":   assoc.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("removing application association: %w", err)
	}
	return nil
}
