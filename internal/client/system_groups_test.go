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

func TestCreateSystemGroup(t *testing.T) {
	var gotBody SystemGroup
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/systemgroups" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := gotBody
		resp.ID = "sg-new"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	group, err := c.CreateSystemGroup(context.Background(), &SystemGroup{
		Name:             "Production Servers",
		Description:      "All production systems",
		MembershipMethod: "STATIC",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != "sg-new" {
		t.Errorf("expected id 'sg-new', got %q", group.ID)
	}
	if gotBody.Name != "Production Servers" {
		t.Errorf("expected name 'Production Servers', got %q", gotBody.Name)
	}
}

func TestGetSystemGroup(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/systemgroups/sg-123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(SystemGroup{
			ID:               "sg-123",
			Name:             "Dev Machines",
			Description:      "Developer workstations",
			MembershipMethod: "DYNAMIC_AUTOMATED",
		})
	}))

	group, err := c.GetSystemGroup(context.Background(), "sg-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != "sg-123" {
		t.Errorf("expected id 'sg-123', got %q", group.ID)
	}
	if group.Name != "Dev Machines" {
		t.Errorf("expected name 'Dev Machines', got %q", group.Name)
	}
	if group.MembershipMethod != "DYNAMIC_AUTOMATED" {
		t.Errorf("expected membership_method 'DYNAMIC_AUTOMATED', got %q", group.MembershipMethod)
	}
}

func TestUpdateSystemGroup(t *testing.T) {
	var gotBody SystemGroup
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/systemgroups/sg-123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := gotBody
		resp.ID = "sg-123"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	group, err := c.UpdateSystemGroup(context.Background(), "sg-123", &SystemGroup{
		Name:        "Updated Name",
		Description: "Updated desc",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %q", group.Name)
	}
}

func TestDeleteSystemGroup(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/systemgroups/sg-del" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))

	err := c.DeleteSystemGroup(context.Background(), "sg-del")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListSystemGroups(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "50" {
			t.Errorf("expected limit=50, got %q", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode([]SystemGroup{
			{ID: "1", Name: "Group A"},
			{ID: "2", Name: "Group B"},
		})
	}))

	groups, err := c.ListSystemGroups(context.Background(), &PaginationParams{Limit: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestGetSystemGroupByName(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]SystemGroup{
			{ID: "1", Name: "Alpha"},
			{ID: "2", Name: "Beta"},
			{ID: "3", Name: "Gamma"},
		})
	}))

	group, err := c.GetSystemGroupByName(context.Background(), "Beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != "2" {
		t.Errorf("expected id '2', got %q", group.ID)
	}
}

func TestGetSystemGroupByName_NotFound(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]SystemGroup{})
	}))

	_, err := c.GetSystemGroupByName(context.Background(), "Missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetSystemGroupByName_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			groups := make([]SystemGroup, 100)
			for i := range groups {
				groups[i] = SystemGroup{ID: "x", Name: "not-it"}
			}
			_ = json.NewEncoder(w).Encode(groups)
		} else {
			_ = json.NewEncoder(w).Encode([]SystemGroup{
				{ID: "found", Name: "Target"},
			})
		}
	}))
	defer server.Close()

	c := New("key", "org", server.URL)
	group, err := c.GetSystemGroupByName(context.Background(), "Target")
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

func TestAddSystemGroupMember(t *testing.T) {
	var gotBody map[string]any
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/systemgroups/sg-1/members" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.AddSystemGroupMember(context.Background(), "sg-1", &SystemGroupMembership{
		ID:   "sys-abc",
		Type: "system",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["op"] != "add" {
		t.Errorf("expected op 'add', got %q", gotBody["op"])
	}
	if gotBody["type"] != "system" {
		t.Errorf("expected type 'system', got %q", gotBody["type"])
	}
	if gotBody["id"] != "sys-abc" {
		t.Errorf("expected id 'sys-abc', got %q", gotBody["id"])
	}
}

func TestRemoveSystemGroupMember(t *testing.T) {
	var gotBody map[string]any
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/systemgroups/sg-1/members" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.RemoveSystemGroupMember(context.Background(), "sg-1", "sys-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["op"] != "remove" {
		t.Errorf("expected op 'remove', got %q", gotBody["op"])
	}
}

func TestGetSystemGroupMembers(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/systemgroups/sg-1/members" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]SystemGroupMembership{
			{ID: "sys-1", Type: "system"},
			{ID: "sys-2", Type: "system"},
		})
	}))

	members, err := c.GetSystemGroupMembers(context.Background(), "sg-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 2 {
		t.Errorf("expected 2 members, got %d", len(members))
	}
}
