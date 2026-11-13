// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type PolicyGroup struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type PolicyGroupMembership struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type PolicyGroupAssociation struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (c *Client) CreatePolicyGroup(ctx context.Context, group *PolicyGroup) (*PolicyGroup, error) {
	var result PolicyGroup
	err := c.do(ctx, "POST", "/policygroups", group, &result)
	if err != nil {
		return nil, fmt.Errorf("creating policy group: %w", err)
	}
	return &result, nil
}

func (c *Client) GetPolicyGroup(ctx context.Context, id string) (*PolicyGroup, error) {
	var result PolicyGroup
	err := c.do(ctx, "GET", "/policygroups/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting policy group: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdatePolicyGroup(ctx context.Context, id string, group *PolicyGroup) (*PolicyGroup, error) {
	var result PolicyGroup
	err := c.do(ctx, "PUT", "/policygroups/"+id, group, &result)
	if err != nil {
		return nil, fmt.Errorf("updating policy group: %w", err)
	}
	return &result, nil
}

func (c *Client) DeletePolicyGroup(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/policygroups/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting policy group: %w", err)
	}
	return nil
}

func (c *Client) ListPolicyGroups(ctx context.Context, params *PaginationParams) ([]PolicyGroup, error) {
	path := c.buildPaginatedPath("/policygroups", params)
	var result []PolicyGroup
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing policy groups: %w", err)
	}
	return result, nil
}

func (c *Client) GetPolicyGroupByName(ctx context.Context, name string) (*PolicyGroup, error) {
	limit := 100
	skip := 0
	for {
		groups, err := c.ListPolicyGroups(ctx, &PaginationParams{Limit: limit, Skip: skip})
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
	return nil, fmt.Errorf("policy group with name %q not found", name)
}

func (c *Client) GetPolicyGroupMembers(ctx context.Context, groupID string) ([]PolicyGroupMembership, error) {
	path := fmt.Sprintf("/policygroups/%s/members", groupID)
	var result []PolicyGroupMembership
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting policy group members: %w", err)
	}
	return result, nil
}

func (c *Client) AddPolicyGroupMember(ctx context.Context, groupID string, member *PolicyGroupMembership) error {
	path := fmt.Sprintf("/policygroups/%s/members", groupID)
	body := map[string]any{
		"op":   "add",
		"type": member.Type,
		"id":   member.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("adding policy group member: %w", err)
	}
	return nil
}

func (c *Client) RemovePolicyGroupMember(ctx context.Context, groupID, memberID string) error {
	path := fmt.Sprintf("/policygroups/%s/members", groupID)
	body := map[string]any{
		"op":   "remove",
		"type": "policy",
		"id":   memberID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("removing policy group member: %w", err)
	}
	return nil
}

func (c *Client) GetPolicyGroupAssociations(ctx context.Context, groupID, targetType string) ([]PolicyGroupAssociation, error) {
	path := fmt.Sprintf("/policygroups/%s/associations?targets=%s", groupID, targetType)
	var result []PolicyGroupAssociation
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting policy group associations: %w", err)
	}
	return result, nil
}

func (c *Client) AddPolicyGroupAssociation(ctx context.Context, groupID string, assoc *PolicyGroupAssociation) error {
	path := fmt.Sprintf("/policygroups/%s/associations", groupID)
	body := map[string]any{
		"op":   "add",
		"type": assoc.Type,
		"id":   assoc.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("adding policy group association: %w", err)
	}
	return nil
}

func (c *Client) RemovePolicyGroupAssociation(ctx context.Context, groupID string, assoc *PolicyGroupAssociation) error {
	path := fmt.Sprintf("/policygroups/%s/associations", groupID)
	body := map[string]any{
		"op":   "remove",
		"type": assoc.Type,
		"id":   assoc.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("removing policy group association: %w", err)
	}
	return nil
}
