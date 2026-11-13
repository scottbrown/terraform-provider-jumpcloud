// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Role struct {
	ID          string           `json:"id,omitempty"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Permissions []RolePermission `json:"permissions"`
	Scope       *RoleScope       `json:"scope,omitempty"`
}

type RolePermission struct {
	ResourceType string `json:"resourceType"`
	Action       string `json:"action"`
}

type RoleScope struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
}

func (r *Role) UnmarshalJSON(data []byte) error {
	// The JumpCloud API returns "scopes" (flat string list) rather than
	// "permissions" (structured objects). Normalize scopes into permissions
	// by splitting on the last dot: "applications.readonly" becomes
	// {resource_type: "applications", action: "readonly"}.
	// Scopes without a dot use action "manage".
	type plain Role
	if err := json.Unmarshal(data, (*plain)(r)); err != nil {
		return err
	}

	// If permissions were already populated (e.g. from a "permissions" field), keep them.
	if len(r.Permissions) > 0 {
		return nil
	}

	var raw struct {
		Scopes []string `json:"scopes"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parsing role scopes: %w", err)
	}

	for _, scope := range raw.Scopes {
		var perm RolePermission
		if idx := strings.LastIndex(scope, "."); idx > 0 {
			perm.ResourceType = scope[:idx]
			perm.Action = scope[idx+1:]
		} else {
			perm.ResourceType = scope
			perm.Action = "manage"
		}
		r.Permissions = append(r.Permissions, perm)
	}

	return nil
}

func (r *Role) MarshalJSON() ([]byte, error) {
	// Convert permissions back to flat scopes for the API.
	scopes := make([]string, 0, len(r.Permissions))
	for _, p := range r.Permissions {
		if p.Action == "manage" {
			scopes = append(scopes, p.ResourceType)
		} else {
			scopes = append(scopes, p.ResourceType+"."+p.Action)
		}
	}

	type plain struct {
		ID          string     `json:"id,omitempty"`
		Name        string     `json:"name"`
		Description string     `json:"description,omitempty"`
		Scopes      []string   `json:"scopes"`
		Scope       *RoleScope `json:"scope,omitempty"`
	}

	return json.Marshal(&plain{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Scopes:      scopes,
		Scope:       r.Scope,
	})
}

func (c *Client) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	var result Role
	err := c.do(ctx, "POST", "/roles", role, &result)
	if err != nil {
		return nil, fmt.Errorf("creating role: %w", err)
	}
	return &result, nil
}

func (c *Client) GetRole(ctx context.Context, id string) (*Role, error) {
	var result Role
	err := c.do(ctx, "GET", "/roles/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting role: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateRole(ctx context.Context, id string, role *Role) (*Role, error) {
	var result Role
	err := c.do(ctx, "PUT", "/roles/"+id, role, &result)
	if err != nil {
		return nil, fmt.Errorf("updating role: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteRole(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/roles/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting role: %w", err)
	}
	return nil
}

func (c *Client) ListRoles(ctx context.Context, params *PaginationParams) ([]Role, error) {
	path := c.buildPaginatedPath("/roles", params)
	var result []Role
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing roles: %w", err)
	}
	return result, nil
}
