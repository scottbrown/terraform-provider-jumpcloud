// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type SoftwareApp struct {
	ID          string          `json:"id,omitempty"`
	DisplayName string          `json:"displayName"`
	Settings    json.RawMessage `json:"settings,omitempty"`
}

type SoftwareAppAssociation struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (c *Client) CreateSoftwareApp(ctx context.Context, app *SoftwareApp) (*SoftwareApp, error) {
	var result SoftwareApp
	err := c.do(ctx, "POST", "/softwareapps", app, &result)
	if err != nil {
		return nil, fmt.Errorf("creating software app: %w", err)
	}
	return &result, nil
}

func (c *Client) GetSoftwareApp(ctx context.Context, id string) (*SoftwareApp, error) {
	var result SoftwareApp
	err := c.do(ctx, "GET", "/softwareapps/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting software app: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateSoftwareApp(ctx context.Context, id string, app *SoftwareApp) (*SoftwareApp, error) {
	var result SoftwareApp
	err := c.do(ctx, "PUT", "/softwareapps/"+id, app, &result)
	if err != nil {
		return nil, fmt.Errorf("updating software app: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteSoftwareApp(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/softwareapps/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting software app: %w", err)
	}
	return nil
}

func (c *Client) ListSoftwareApps(ctx context.Context, params *PaginationParams) ([]SoftwareApp, error) {
	path := c.buildPaginatedPath("/softwareapps", params)
	var result []SoftwareApp
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing software apps: %w", err)
	}
	return result, nil
}

func (c *Client) GetSoftwareAppByName(ctx context.Context, name string) (*SoftwareApp, error) {
	limit := 100
	skip := 0
	for {
		apps, err := c.ListSoftwareApps(ctx, &PaginationParams{Limit: limit, Skip: skip})
		if err != nil {
			return nil, err
		}
		for i := range apps {
			if apps[i].DisplayName == name {
				return &apps[i], nil
			}
		}
		if len(apps) < limit {
			break
		}
		skip += limit
	}
	return nil, fmt.Errorf("software app with name %q not found", name)
}

func (c *Client) GetSoftwareAppAssociations(ctx context.Context, appID, targetType string) ([]SoftwareAppAssociation, error) {
	path := fmt.Sprintf("/softwareapps/%s/associations?targets=%s", appID, targetType)
	var result []SoftwareAppAssociation
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting software app associations: %w", err)
	}
	return result, nil
}

func (c *Client) AddSoftwareAppAssociation(ctx context.Context, appID string, assoc *SoftwareAppAssociation) error {
	path := fmt.Sprintf("/softwareapps/%s/associations", appID)
	body := map[string]any{
		"op":   "add",
		"type": assoc.Type,
		"id":   assoc.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("adding software app association: %w", err)
	}
	return nil
}

func (c *Client) RemoveSoftwareAppAssociation(ctx context.Context, appID string, assoc *SoftwareAppAssociation) error {
	path := fmt.Sprintf("/softwareapps/%s/associations", appID)
	body := map[string]any{
		"op":   "remove",
		"type": assoc.Type,
		"id":   assoc.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("removing software app association: %w", err)
	}
	return nil
}
