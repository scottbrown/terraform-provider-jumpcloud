// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"fmt"
)

type UserGroup struct {
	ID               string                `json:"id,omitempty"`
	Name             string                `json:"name"`
	Description      string                `json:"description,omitempty"`
	MembershipMethod string                `json:"membershipMethod,omitempty"`
	Attributes       map[string]any        `json:"attributes,omitempty"`
	MemberQuery      *UserGroupMemberQuery `json:"memberQuery,omitempty"`
}

type UserGroupMemberQuery struct {
	QueryType string              `json:"queryType,omitempty"`
	Filters   []MemberQueryFilter `json:"filters,omitempty"`
}

type MemberQueryFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type UserGroupMembership struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func (c *Client) CreateUserGroup(ctx context.Context, group *UserGroup) (*UserGroup, error) {
	var result UserGroup
	err := c.do(ctx, "POST", "/usergroups", group, &result)
	if err != nil {
		return nil, fmt.Errorf("creating user group: %w", err)
	}
	return &result, nil
}

func (c *Client) GetUserGroup(ctx context.Context, id string) (*UserGroup, error) {
	var result UserGroup
	err := c.do(ctx, "GET", "/usergroups/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting user group: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateUserGroup(ctx context.Context, id string, group *UserGroup) (*UserGroup, error) {
	var result UserGroup
	err := c.do(ctx, "PUT", "/usergroups/"+id, group, &result)
	if err != nil {
		return nil, fmt.Errorf("updating user group: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteUserGroup(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/usergroups/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting user group: %w", err)
	}
	return nil
}

func (c *Client) ListUserGroups(ctx context.Context, params *PaginationParams) ([]UserGroup, error) {
	path := c.buildPaginatedPath("/usergroups", params)
	var result []UserGroup
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing user groups: %w", err)
	}
	return result, nil
}

func (c *Client) GetUserGroupByName(ctx context.Context, name string) (*UserGroup, error) {
	limit := 100
	skip := 0
	for {
		groups, err := c.ListUserGroups(ctx, &PaginationParams{Limit: limit, Skip: skip})
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
	return nil, fmt.Errorf("user group with name %q not found", name)
}

func (c *Client) GetUserGroupMembers(ctx context.Context, groupID string) ([]UserGroupMembership, error) {
	path := fmt.Sprintf("/usergroups/%s/members", groupID)
	var result []UserGroupMembership
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting user group members: %w", err)
	}
	return result, nil
}

func (c *Client) AddUserGroupMember(ctx context.Context, groupID string, member *UserGroupMembership) error {
	path := fmt.Sprintf("/usergroups/%s/members", groupID)
	body := map[string]any{
		"op":   "add",
		"type": member.Type,
		"id":   member.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("adding user group member: %w", err)
	}
	return nil
}

func (c *Client) RemoveUserGroupMember(ctx context.Context, groupID, memberID string) error {
	path := fmt.Sprintf("/usergroups/%s/members", groupID)
	body := map[string]any{
		"op":   "remove",
		"type": "user",
		"id":   memberID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("removing user group member: %w", err)
	}
	return nil
}
