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

func TestGetPolicyGroup(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/policygroups/pg123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(PolicyGroup{
			ID:          "pg123",
			Name:        "Compliance Policies",
			Description: "All compliance-related policies",
		})
	}))

	group, err := c.GetPolicyGroup(context.Background(), "pg123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != "pg123" {
		t.Errorf("expected id 'pg123', got %q", group.ID)
	}
	if group.Name != "Compliance Policies" {
		t.Errorf("expected name 'Compliance Policies', got %q", group.Name)
	}
}

func TestCreatePolicyGroup(t *testing.T) {
	var gotBody PolicyGroup
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/policygroups" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := gotBody
		resp.ID = "new-id"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	group, err := c.CreatePolicyGroup(context.Background(), &PolicyGroup{Name: "New Group"})
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

func TestDeletePolicyGroup(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/policygroups/del-id" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeletePolicyGroup(context.Background(), "del-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListPolicyGroups(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %q", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode([]PolicyGroup{
			{ID: "1", Name: "Group A"},
			{ID: "2", Name: "Group B"},
		})
	}))

	groups, err := c.ListPolicyGroups(context.Background(), &PaginationParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
}

func TestGetPolicyGroupByName(t *testing.T) {
	groups := []PolicyGroup{
		{ID: "1", Name: "Alpha"},
		{ID: "2", Name: "Beta"},
		{ID: "3", Name: "Gamma"},
	}
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(groups)
	}))

	group, err := c.GetPolicyGroupByName(context.Background(), "Beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != "2" {
		t.Errorf("expected id '2', got %q", group.ID)
	}
}

func TestGetPolicyGroupByName_NotFound(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode([]PolicyGroup{})
	}))

	_, err := c.GetPolicyGroupByName(context.Background(), "Missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPolicyGroupByName_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			groups := make([]PolicyGroup, 100)
			for i := range groups {
				groups[i] = PolicyGroup{ID: "x", Name: "not-it"}
			}
			_ = json.NewEncoder(w).Encode(groups)
		} else {
			_ = json.NewEncoder(w).Encode([]PolicyGroup{
				{ID: "found", Name: "Target"},
			})
		}
	}))
	defer server.Close()

	c := New("key", "org", server.URL)
	group, err := c.GetPolicyGroupByName(context.Background(), "Target")
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

func TestAddPolicyGroupMember(t *testing.T) {
	var gotBody map[string]any
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/policygroups/pg1/members" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.AddPolicyGroupMember(context.Background(), "pg1", &PolicyGroupMembership{
		ID:   "policy-1",
		Type: "policy",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["op"] != "add" {
		t.Errorf("expected op 'add', got %q", gotBody["op"])
	}
	if gotBody["id"] != "policy-1" {
		t.Errorf("expected id 'policy-1', got %q", gotBody["id"])
	}
}

func TestRemovePolicyGroupMember(t *testing.T) {
	var gotBody map[string]any
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/policygroups/pg1/members" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.RemovePolicyGroupMember(context.Background(), "pg1", "policy-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["op"] != "remove" {
		t.Errorf("expected op 'remove', got %q", gotBody["op"])
	}
}

func TestAddPolicyGroupAssociation(t *testing.T) {
	var gotBody map[string]any
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/policygroups/pg1/associations" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.AddPolicyGroupAssociation(context.Background(), "pg1", &PolicyGroupAssociation{
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
