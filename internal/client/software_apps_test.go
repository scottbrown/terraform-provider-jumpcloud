// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSoftwareApp(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/softwareapps/app123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(SoftwareApp{
			ID:          "app123",
			DisplayName: "Google Chrome",
			Settings:    json.RawMessage(`{"packageId":"com.google.chrome"}`),
		})
	}))

	app, err := c.GetSoftwareApp(context.Background(), "app123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ID != "app123" {
		t.Errorf("expected id 'app123', got %q", app.ID)
	}
	if app.DisplayName != "Google Chrome" {
		t.Errorf("expected display_name 'Google Chrome', got %q", app.DisplayName)
	}
}

func TestCreateSoftwareApp(t *testing.T) {
	var gotBody SoftwareApp
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/softwareapps" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := gotBody
		resp.ID = "new-id"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	app, err := c.CreateSoftwareApp(context.Background(), &SoftwareApp{
		DisplayName: "Slack",
		Settings:    json.RawMessage(`{"packageId":"com.slack"}`),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ID != "new-id" {
		t.Errorf("expected id 'new-id', got %q", app.ID)
	}
	if gotBody.DisplayName != "Slack" {
		t.Errorf("expected request body display_name 'Slack', got %q", gotBody.DisplayName)
	}
}

func TestDeleteSoftwareApp(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/softwareapps/del-id" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeleteSoftwareApp(context.Background(), "del-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListSoftwareApps(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %q", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode([]SoftwareApp{
			{ID: "1", DisplayName: "Chrome"},
			{ID: "2", DisplayName: "Firefox"},
		})
	}))

	apps, err := c.ListSoftwareApps(context.Background(), &PaginationParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 2 {
		t.Errorf("expected 2 apps, got %d", len(apps))
	}
}

func TestGetSoftwareAppByName(t *testing.T) {
	apps := []SoftwareApp{
		{ID: "1", DisplayName: "Chrome"},
		{ID: "2", DisplayName: "Slack"},
		{ID: "3", DisplayName: "Zoom"},
	}
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(apps)
	}))

	app, err := c.GetSoftwareAppByName(context.Background(), "Slack")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ID != "2" {
		t.Errorf("expected id '2', got %q", app.ID)
	}
}

func TestGetSoftwareAppByName_NotFound(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]SoftwareApp{})
	}))

	_, err := c.GetSoftwareAppByName(context.Background(), "Missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetSoftwareAppByName_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			apps := make([]SoftwareApp, 100)
			for i := range apps {
				apps[i] = SoftwareApp{ID: "x", DisplayName: "not-it"}
			}
			_ = json.NewEncoder(w).Encode(apps)
		} else {
			_ = json.NewEncoder(w).Encode([]SoftwareApp{
				{ID: "found", DisplayName: "Target"},
			})
		}
	}))
	defer server.Close()

	c := New("key", "org", server.URL)
	app, err := c.GetSoftwareAppByName(context.Background(), "Target")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.ID != "found" {
		t.Errorf("expected id 'found', got %q", app.ID)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}

func TestAddSoftwareAppAssociation(t *testing.T) {
	var gotBody map[string]any
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/softwareapps/app1/associations" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.AddSoftwareAppAssociation(context.Background(), "app1", &SoftwareAppAssociation{
		ID:   "sys-group-1",
		Type: "system_group",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["op"] != "add" {
		t.Errorf("expected op 'add', got %q", gotBody["op"])
	}
	if gotBody["type"] != "system_group" {
		t.Errorf("expected type 'system_group', got %q", gotBody["type"])
	}
}

func TestRemoveSoftwareAppAssociation(t *testing.T) {
	var gotBody map[string]any
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/softwareapps/app1/associations" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.RemoveSoftwareAppAssociation(context.Background(), "app1", &SoftwareAppAssociation{
		ID:   "sys-1",
		Type: "system",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["op"] != "remove" {
		t.Errorf("expected op 'remove', got %q", gotBody["op"])
	}
}
