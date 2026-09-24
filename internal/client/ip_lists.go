// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"fmt"
)

type IPList struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	IPs         []string `json:"ips"`
}

func (c *Client) CreateIPList(ctx context.Context, ipList *IPList) (*IPList, error) {
	var result IPList
	err := c.do(ctx, "POST", "/iplists", ipList, &result)
	if err != nil {
		return nil, fmt.Errorf("creating ip list: %w", err)
	}
	return &result, nil
}

func (c *Client) GetIPList(ctx context.Context, id string) (*IPList, error) {
	var result IPList
	err := c.do(ctx, "GET", "/iplists/"+id, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("getting ip list: %w", err)
	}
	return &result, nil
}

func (c *Client) UpdateIPList(ctx context.Context, id string, ipList *IPList) (*IPList, error) {
	var result IPList
	err := c.do(ctx, "PUT", "/iplists/"+id, ipList, &result)
	if err != nil {
		return nil, fmt.Errorf("updating ip list: %w", err)
	}
	return &result, nil
}

func (c *Client) DeleteIPList(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/iplists/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting ip list: %w", err)
	}
	return nil
}

func (c *Client) ListIPLists(ctx context.Context, params *PaginationParams) ([]IPList, error) {
	path := c.buildPaginatedPath("/iplists", params)
	var result []IPList
	err := c.do(ctx, "GET", path, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("listing ip lists: %w", err)
	}
	return result, nil
}
