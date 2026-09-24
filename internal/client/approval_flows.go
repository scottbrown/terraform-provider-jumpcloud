// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type ApprovalFlow struct {
	ID                  string                 `json:"id,omitempty"`
	Name                string                 `json:"name"`
	Description         string                 `json:"description,omitempty"`
	ApprovalType        string                 `json:"approvalType,omitempty"`
	ApproverRequirement string                 `json:"approverRequirement,omitempty"`
	ApproverResources   []ApproverResource     `json:"approverResources,omitempty"`
	Category            string                 `json:"category,omitempty"`
	DeviceLevelPerm     json.RawMessage        `json:"deviceLevelPermission,omitempty"`
	FixedDuration       string                 `json:"fixedDuration,omitempty"`
	IconColor           string                 `json:"iconColor,omitempty"`
	IconURL             string                 `json:"iconUrl,omitempty"`
	MultiSelectDuration []string               `json:"multiSelectDuration,omitempty"`
	NonAdminApproval    bool                   `json:"nonAdminApproval,omitempty"`
	OrgID               string                 `json:"organizationObjectId,omitempty"`
	Resources           []ApprovalFlowResource `json:"resources,omitempty"`
	ResourceID          string                 `json:"resourceId,omitempty"`
	ResourceType        string                 `json:"resourceType,omitempty"`
	SlackConfig         json.RawMessage        `json:"slackConfig,omitempty"`
	SlackDMEnabled      bool                   `json:"slackDmEnabled,omitempty"`
	SlackEnabled        bool                   `json:"slackEnabled,omitempty"`
	Status              string                 `json:"status,omitempty"`
	TimeBasedAccess     bool                   `json:"timeBasedAccess,omitempty"`
	TTLConfig           string                 `json:"ttlConfig,omitempty"`
	VisibleTo           []VisibleToEntry       `json:"visibleTo,omitempty"`
}

type ApprovalFlowResource struct {
	ResourceID   string `json:"resourceId"`
	ResourceType string `json:"resourceType"`
	ResourceName string `json:"resourceName,omitempty"`
}

type VisibleToEntry struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type ApproverResource struct {
	ResourceID    string `json:"resourceId"`
	ResourceType  string `json:"resourceType"`
	ResourceName  string `json:"resourceName,omitempty"`
	ResourceOrder int    `json:"resourceOrder,omitempty"`
	RequireAll    bool   `json:"requireAll,omitempty"`
}

// GetResourceID returns the resource ID from the Resources array or the flat field.
func (f *ApprovalFlow) GetResourceID() string {
	if len(f.Resources) > 0 {
		return f.Resources[0].ResourceID
	}
	return f.ResourceID
}

// GetResourceType returns the resource type from the Resources array or the flat field.
func (f *ApprovalFlow) GetResourceType() string {
	if len(f.Resources) > 0 {
		return f.Resources[0].ResourceType
	}
	return f.ResourceType
}

// GetVisibleToIDs returns just the IDs from the VisibleTo entries.
func (f *ApprovalFlow) GetVisibleToIDs() []string {
	ids := make([]string, len(f.VisibleTo))
	for i, v := range f.VisibleTo {
		ids[i] = v.ID
	}
	return ids
}

type ApprovalFlowSettings struct {
	ID                     string         `json:"id,omitempty"`
	AdminApprovalEmail     bool           `json:"adminApprovalEmail,omitempty"`
	ChannelEvents          []ChannelEvent `json:"channelEvents,omitempty"`
	ExposeApprovalProgress bool           `json:"exposeApprovalProgress,omitempty"`
	OrgID                  string         `json:"organizationObjectId,omitempty"`
	ResourceRequest        bool           `json:"resourceRequest,omitempty"`
	UserApprovalEmail      bool           `json:"userApprovalEmail,omitempty"`
	UserRequestEmail       bool           `json:"userRequestEmail,omitempty"`
}

type ChannelEvent struct {
	Channel string   `json:"channel,omitempty"`
	Events  []string `json:"events,omitempty"`
}

func (c *Client) CreateApprovalFlow(ctx context.Context, flow *ApprovalFlow) (*ApprovalFlow, error) {
	var result ApprovalFlow
	err := c.do(ctx, "POST", "/approvalflows", flow, &result)
	if err != nil {
		return nil, fmt.Errorf("creating approval flow: %w", err)
	}
	return &result, nil
}

func (c *Client) GetApprovalFlow(ctx context.Context, id string) (*ApprovalFlow, error) {
	var result ApprovalFlow
	err := c.do(ctx, "GET", "/approvalflows/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting approval flow: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateApprovalFlow(ctx context.Context, id string, flow *ApprovalFlow) (*ApprovalFlow, error) {
	var result ApprovalFlow
	err := c.do(ctx, "PUT", "/approvalflows/"+id, flow, &result)
	if err != nil {
		return nil, fmt.Errorf("updating approval flow: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteApprovalFlow(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/approvalflows/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting approval flow: %w", err)
	}
	return nil
}

func (c *Client) ListApprovalFlows(ctx context.Context, params *PaginationParams) ([]ApprovalFlow, error) {
	path := c.buildPaginatedPath("/approvalflows", params)
	var result []ApprovalFlow
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing approval flows: %w", err)
	}
	return result, nil
}

func (c *Client) GetApprovalFlowSettings(ctx context.Context) (*ApprovalFlowSettings, error) {
	var result ApprovalFlowSettings
	err := c.do(ctx, "GET", "/approvalflowsettings", nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting approval flow settings: %w", err)
	}
	return &result, nil
}

func (c *Client) CreateApprovalFlowSettings(ctx context.Context, settings *ApprovalFlowSettings) (*ApprovalFlowSettings, error) {
	var result ApprovalFlowSettings
	err := c.do(ctx, "POST", "/approvalflowsettings", settings, &result)
	if err != nil {
		return nil, fmt.Errorf("creating approval flow settings: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateApprovalFlowSettings(ctx context.Context, settings *ApprovalFlowSettings) (*ApprovalFlowSettings, error) {
	var result ApprovalFlowSettings
	err := c.do(ctx, "PUT", "/approvalflowsettings", settings, &result)
	if err != nil {
		return nil, fmt.Errorf("updating approval flow settings: %w", err)
	}
	return &result, nil
}
