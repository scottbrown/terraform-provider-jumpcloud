// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
)

type NotificationChannel struct {
	ID          string          `json:"objectId,omitempty"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Enabled     bool            `json:"enabled,omitempty"`
	Type        string          `json:"type"`
	Config      json.RawMessage `json:"config,omitempty"`
}

type notificationChannelRequest struct {
	Channel *NotificationChannel `json:"channel"`
}

type notificationChannelResponse struct {
	Channel NotificationChannel `json:"channel"`
}

type notificationChannelListResponse struct {
	Channels []NotificationChannel `json:"channels"`
}

func (c *Client) CreateNotificationChannel(ctx context.Context, channel *NotificationChannel) (*NotificationChannel, error) {
	var resp notificationChannelResponse
	body := notificationChannelRequest{Channel: channel}
	err := c.do(ctx, "POST", "/notifications/channels", body, &resp)
	if err != nil {
		return nil, fmt.Errorf("creating notification channel: %w", err)
	}
	return &resp.Channel, nil
}

func (c *Client) GetNotificationChannel(ctx context.Context, id string) (*NotificationChannel, error) {
	var resp notificationChannelResponse
	err := c.do(ctx, "GET", "/notifications/channels/"+id, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("getting notification channel: %w", err)
	}
	return &resp.Channel, nil
}

func (c *Client) UpdateNotificationChannel(ctx context.Context, id string, channel *NotificationChannel) (*NotificationChannel, error) {
	var resp notificationChannelResponse
	body := notificationChannelRequest{Channel: channel}
	err := c.do(ctx, "PATCH", "/notifications/channels/"+id, body, &resp)
	if err != nil {
		return nil, fmt.Errorf("updating notification channel: %w", err)
	}
	return &resp.Channel, nil
}

func (c *Client) DeleteNotificationChannel(ctx context.Context, id string) error {
	err := c.do(ctx, "DELETE", "/notifications/channels/"+id, nil, nil)
	if err != nil {
		return fmt.Errorf("deleting notification channel: %w", err)
	}
	return nil
}

func (c *Client) ListNotificationChannels(ctx context.Context, params *PaginationParams) ([]NotificationChannel, error) {
	path := c.buildPaginatedPath("/notifications/channels", params)
	var resp notificationChannelListResponse
	err := c.do(ctx, "GET", path, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("listing notification channels: %w", err)
	}
	return resp.Channels, nil
}

func (c *Client) GetNotificationChannelByName(ctx context.Context, name string) (*NotificationChannel, error) {
	limit := 100
	skip := 0
	for {
		channels, err := c.ListNotificationChannels(ctx, &PaginationParams{Limit: limit, Skip: skip})
		if err != nil {
			return nil, err
		}
		for i := range channels {
			if channels[i].Name == name {
				return c.GetNotificationChannel(ctx, channels[i].ID)
			}
		}
		if len(channels) < limit {
			break
		}
		skip += limit
	}
	return nil, fmt.Errorf("notification channel with name %q not found", name)
}
