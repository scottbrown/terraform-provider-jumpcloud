// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestCreateNotificationChannel(t *testing.T) {
	var gotBody notificationChannelRequest
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/notifications/channels" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := notificationChannelResponse{
			Channel: *gotBody.Channel,
		}
		resp.Channel.ID = "ch-new"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	channel, err := c.CreateNotificationChannel(context.Background(), &NotificationChannel{
		Name:    "Production Alerts",
		Type:    "CHANNEL_TYPE_WEBHOOK",
		Enabled: true,
		Config:  json.RawMessage(`{"webhook":{"url":"https://hooks.example.com"}}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channel.ID != "ch-new" {
		t.Errorf("expected id 'ch-new', got %q", channel.ID)
	}
	if gotBody.Channel.Name != "Production Alerts" {
		t.Errorf("expected name 'Production Alerts', got %q", gotBody.Channel.Name)
	}
	if gotBody.Channel.Type != "CHANNEL_TYPE_WEBHOOK" {
		t.Errorf("expected type 'CHANNEL_TYPE_WEBHOOK', got %q", gotBody.Channel.Type)
	}
}

func TestGetNotificationChannel(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/notifications/channels/ch-123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(notificationChannelResponse{
			Channel: NotificationChannel{
				ID:      "ch-123",
				Name:    "Slack Alerts",
				Type:    "CHANNEL_TYPE_SLACK",
				Enabled: true,
			},
		})
	}))

	channel, err := c.GetNotificationChannel(context.Background(), "ch-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channel.ID != "ch-123" {
		t.Errorf("expected id 'ch-123', got %q", channel.ID)
	}
	if channel.Name != "Slack Alerts" {
		t.Errorf("expected name 'Slack Alerts', got %q", channel.Name)
	}
}

func TestUpdateNotificationChannel(t *testing.T) {
	var gotBody notificationChannelRequest
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/notifications/channels/ch-123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := notificationChannelResponse{
			Channel: *gotBody.Channel,
		}
		resp.Channel.ID = "ch-123"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	channel, err := c.UpdateNotificationChannel(context.Background(), "ch-123", &NotificationChannel{
		Name:    "Updated Channel",
		Type:    "CHANNEL_TYPE_WEBHOOK",
		Enabled: false,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channel.Name != "Updated Channel" {
		t.Errorf("expected name 'Updated Channel', got %q", channel.Name)
	}
}

func TestDeleteNotificationChannel(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/notifications/channels/ch-del" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))

	err := c.DeleteNotificationChannel(context.Background(), "ch-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListNotificationChannels(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/notifications/channels" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "50" {
			t.Errorf("expected limit=50, got %q", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode(notificationChannelListResponse{
			Channels: []NotificationChannel{
				{ID: "1", Name: "Channel A"},
				{ID: "2", Name: "Channel B"},
			},
		})
	}))

	channels, err := c.ListNotificationChannels(context.Background(), &PaginationParams{Limit: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(channels) != 2 {
		t.Errorf("expected 2 channels, got %d", len(channels))
	}
}

func TestGetNotificationChannelByName(t *testing.T) {
	callCount := 0
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			_ = json.NewEncoder(w).Encode(notificationChannelListResponse{
				Channels: []NotificationChannel{
					{ID: "1", Name: "Channel A"},
					{ID: "2", Name: "Target Channel"},
				},
			})
		} else {
			_ = json.NewEncoder(w).Encode(notificationChannelResponse{
				Channel: NotificationChannel{
					ID:   "2",
					Name: "Target Channel",
					Type: "CHANNEL_TYPE_SLACK",
				},
			})
		}
	}))

	channel, err := c.GetNotificationChannelByName(context.Background(), "Target Channel")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if channel.ID != "2" {
		t.Errorf("expected id '2', got %q", channel.ID)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls (list + get), got %d", callCount)
	}
}

func TestGetNotificationChannelByName_NotFound(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(notificationChannelListResponse{
			Channels: []NotificationChannel{},
		})
	}))

	_, err := c.GetNotificationChannelByName(context.Background(), "Missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
