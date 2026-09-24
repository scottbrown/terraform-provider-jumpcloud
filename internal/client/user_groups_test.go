// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserGroup(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/usergroups/group123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(UserGroup{
			ID:          "group123",
			Name:        "Test Group",
			Description: "A test group",
		})
	}))

	group, err := c.GetUserGroup(context.Background(), "group123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != "group123" {
		t.Errorf("expected id 'group123', got %q", group.ID)
	}
	if group.Name != "Test Group" {
		t.Errorf("expected name 'Test Group', got %q", group.Name)
	}
}

func TestCreateUserGroup(t *testing.T) {
	var gotBody UserGroup
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/usergroups" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := gotBody
		resp.ID = "new-id"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	group, err := c.CreateUserGroup(context.Background(), &UserGroup{Name: "New Group"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != "new-id" {
		t.Errorf("expected id 'new-id', got %q", group.ID)
	}
	if gotBody.Name != "New Group" {
		t.Errorf("expected request body name 'New Group', got %q", gotBody.Name)
	}
}

func TestDeleteUserGroup(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/usergroups/del-id" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeleteUserGroup(context.Background(), "del-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetUserGroupByName(t *testing.T) {
	groups := []UserGroup{
		{ID: "1", Name: "Alpha"},
		{ID: "2", Name: "Beta"},
		{ID: "3", Name: "Gamma"},
	}
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(groups)
	}))

	group, err := c.GetUserGroupByName(context.Background(), "Beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != "2" {
		t.Errorf("expected id '2', got %q", group.ID)
	}
}

func TestGetUserGroupByName_NotFound(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]UserGroup{})
	}))

	_, err := c.GetUserGroupByName(context.Background(), "Missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListUserGroups(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %q", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode([]UserGroup{
			{ID: "1", Name: "Group A"},
			{ID: "2", Name: "Group B"},
		})
	}))

	groups, err := c.ListUserGroups(context.Background(), &PaginationParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestGetUserGroupByName_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Return full page (100 items) to trigger pagination
			groups := make([]UserGroup, 100)
			for i := range groups {
				groups[i] = UserGroup{ID: "x", Name: "not-it"}
			}
			_ = json.NewEncoder(w).Encode(groups)
		} else {
			// Second page has the target
			_ = json.NewEncoder(w).Encode([]UserGroup{
				{ID: "found", Name: "Target"},
			})
		}
	}))
	defer server.Close()

	c := New("key", "org", server.URL)
	group, err := c.GetUserGroupByName(context.Background(), "Target")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != "found" {
		t.Errorf("expected id 'found', got %q", group.ID)
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
}
