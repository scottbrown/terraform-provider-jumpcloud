// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"fmt"
)

type Command struct {
	ID             string   `json:"id,omitempty"`
	Name           string   `json:"name"`
	Command        string   `json:"command"`
	CommandType    string   `json:"commandType"`
	User           string   `json:"user,omitempty"`
	Sudo           bool     `json:"sudo,omitempty"`
	Shell          string   `json:"shell,omitempty"`
	Trigger        string   `json:"trigger,omitempty"`
	Schedule       string   `json:"schedule,omitempty"`
	ScheduleRepeat string   `json:"scheduleRepeatType,omitempty"`
	Timeout        string   `json:"timeout,omitempty"`
	Files          []string `json:"files,omitempty"`
	LaunchType     string   `json:"launchType,omitempty"`
	Organization   string   `json:"organization,omitempty"`
}

type CommandAssociation struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type commandListResponse struct {
	TotalCount int       `json:"totalCount"`
	Results    []Command `json:"results"`
}

func (c *Client) CreateCommand(ctx context.Context, cmd *Command) (*Command, error) {
	var result Command
	err := c.doV1(ctx, "POST", "/commands", cmd, &result)
	if err != nil {
		return nil, fmt.Errorf("creating command: %w", err)
	}
	return &result, nil
}

func (c *Client) GetCommand(ctx context.Context, id string) (*Command, error) {
	var result Command
	err := c.doV1(ctx, "GET", "/commands/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting command: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateCommand(ctx context.Context, id string, cmd *Command) (*Command, error) {
	var result Command
	err := c.doV1(ctx, "PUT", "/commands/"+id, cmd, &result)
	if err != nil {
		return nil, fmt.Errorf("updating command: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteCommand(ctx context.Context, id string) error {
	err := c.doV1(ctx, "DELETE", "/commands/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting command: %w", err)
	}
	return nil
}

func (c *Client) ListCommands(ctx context.Context, params *PaginationParams) ([]Command, error) {
	path := c.buildPaginatedPath("/commands", params)
	var resp commandListResponse
	err := c.doV1(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("listing commands: %w", err)
	}
	return resp.Results, nil
}

func (c *Client) GetCommandAssociations(ctx context.Context, commandID, targetType string) ([]CommandAssociation, error) {
	path := fmt.Sprintf("/commands/%s/associations?targets=%s", commandID, targetType)
	var result []CommandAssociation
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting command associations: %w", err)
	}
	return result, nil
}

func (c *Client) AddCommandAssociation(ctx context.Context, commandID string, assoc *CommandAssociation) error {
	path := fmt.Sprintf("/commands/%s/associations", commandID)
	body := map[string]any{
		"op":   "add",
		"type": assoc.Type,
		"id":   assoc.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("adding command association: %w", err)
	}
	return nil
}

func (c *Client) RemoveCommandAssociation(ctx context.Context, commandID string, assoc *CommandAssociation) error {
	path := fmt.Sprintf("/commands/%s/associations", commandID)
	body := map[string]any{
		"op":   "remove",
		"type": assoc.Type,
		"id":   assoc.ID,
	}
	err := c.do(ctx, "POST", path, body, nil)
	if err != nil {
		return fmt.Errorf("removing command association: %w", err)
	}
	return nil
}
