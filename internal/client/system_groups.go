// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"fmt"
)

type SystemGroup struct {
	ID                      string                  `json:"id,omitempty"`
	Name                    string                  `json:"name"`
	Description             string                  `json:"description,omitempty"`
	MembershipMethod        string                  `json:"membershipMethod,omitempty"`
	MemberQuery             *SystemGroupMemberQuery `json:"memberQuery,omitempty"`
	MemberSuggestionsNotify bool                    `json:"memberSuggestionsNotify,omitempty"`
}

type SystemGroupMemberQuery struct {
	QueryType     string   `json:"queryType,omitempty"`
	Filters       []string `json:"filters,omitempty"`
	SearchFilters string   `json:"searchFilters,omitempty"`
}

type SystemGroupMembership struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (c *Client) CreateSystemGroup(ctx context.Context, group *SystemGroup) (*SystemGroup, error) {
	var result SystemGroup
	err := c.do(ctx, "POST", "/systemgroups", group, &result)
	if err != nil {
		return nil, fmt.Errorf("creating system group: %w", err)
	}
	return &result, nil
}

func (c *Client) GetSystemGroup(ctx context.Context, id string) (*SystemGroup, error) {
	var result SystemGroup
	err := c.do(ctx, "GET", "/systemgroups/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting system group: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateSystemGroup(ctx context.Context, id string, group *SystemGroup) (*SystemGroup, error) {
	var result SystemGroup
	err := c.do(ctx, "PUT", "/systemgroups/"+id, group, &result)
	if err != nil {
		return nil, fmt.Errorf("updating system group: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteSystemGroup(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/systemgroups/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting system group: %w", err)
	}
	return nil
}

func (c *Client) ListSystemGroups(ctx context.Context, params *PaginationParams) ([]SystemGroup, error) {
	path := c.buildPaginatedPath("/systemgroups", params)
	var result []SystemGroup
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing system groups: %w", err)
	}
	return result, nil
}

func (c *Client) GetSystemGroupByName(ctx context.Context, name string) (*SystemGroup, error) {
	limit := 100
	skip := 0
	for {
		groups, err := c.ListSystemGroups(ctx, &PaginationParams{Limit: limit, Skip: skip})
		if err != nil {
			return nil, err
		}
		for i := range groups {
			if groups[i].Name == name {
				return &groups[i], nil
			}
		}
		if len(groups) < limit {
			break
		}
		skip += limit
	}
	return nil, fmt.Errorf("system group with name %q not found", name)
}

func (c *Client) GetSystemGroupMembers(ctx context.Context, groupID string) ([]SystemGroupMembership, error) {
	path := fmt.Sprintf("/systemgroups/%s/members", groupID)
	var result []SystemGroupMembership
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting system group members: %w", err)
	}
	return result, nil
}

func (c *Client) AddSystemGroupMember(ctx context.Context, groupID string, member *SystemGroupMembership) error {
	path := fmt.Sprintf("/systemgroups/%s/members", groupID)
	body := map[string]any{
		"op":   "add",
		"type": member.Type,
		"id":   member.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("adding system group member: %w", err)
	}
	return nil
}

func (c *Client) RemoveSystemGroupMember(ctx context.Context, groupID, memberID string) error {
	path := fmt.Sprintf("/systemgroups/%s/members", groupID)
	body := map[string]any{
		"op":   "remove",
		"type": "system",
		"id":   memberID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("removing system group member: %w", err)
	}
	return nil
}
