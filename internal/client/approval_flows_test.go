// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetApprovalFlow(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/approvalflows/flow123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(ApprovalFlow{
			ID:               "flow123",
			Name:             "Test Flow",
			ApprovalType:     "manual",
			Status:           "active",
			TimeBasedAccess:  true,
			NonAdminApproval: true,
		})
	}))

	flow, err := c.GetApprovalFlow(context.Background(), "flow123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flow.ID != "flow123" {
		t.Errorf("expected id 'flow123', got %q", flow.ID)
	}
	if flow.Name != "Test Flow" {
		t.Errorf("expected name 'Test Flow', got %q", flow.Name)
	}
	if !flow.TimeBasedAccess {
		t.Error("expected TimeBasedAccess to be true")
	}
}

func TestCreateApprovalFlow(t *testing.T) {
	var gotBody ApprovalFlow
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/approvalflows" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := gotBody
		resp.ID = "created-id"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	flow, err := c.CreateApprovalFlow(context.Background(), &ApprovalFlow{
		Name:         "New Flow",
		ApprovalType: "automated",
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flow.ID != "created-id" {
		t.Errorf("expected id 'created-id', got %q", flow.ID)
	}
	if gotBody.Name != "New Flow" {
		t.Errorf("expected request name 'New Flow', got %q", gotBody.Name)
	}
}

func TestUpdateApprovalFlow(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/approvalflows/upd-id" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body ApprovalFlow
		_ = json.NewDecoder(r.Body).Decode(&body)
		body.ID = "upd-id"
		_ = json.NewEncoder(w).Encode(body)
	}))

	flow, err := c.UpdateApprovalFlow(context.Background(), "upd-id", &ApprovalFlow{
		Name:   "Updated",
		Status: "disabled",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flow.ID != "upd-id" {
		t.Errorf("expected id 'upd-id', got %q", flow.ID)
	}
}

func TestDeleteApprovalFlow(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/approvalflows/del-id" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeleteApprovalFlow(context.Background(), "del-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApprovalFlow_GetVisibleToIDs(t *testing.T) {
	flow := ApprovalFlow{
		VisibleTo: []VisibleToEntry{
			{ID: "aaa"},
			{ID: "bbb"},
		},
	}
	ids := flow.GetVisibleToIDs()
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
	if ids[0] != "aaa" || ids[1] != "bbb" {
		t.Errorf("unexpected ids: %v", ids)
	}
}

func TestApprovalFlow_GetResourceID(t *testing.T) {
	flow := ApprovalFlow{
		Resources: []ApprovalFlowResource{
			{ResourceID: "res-1", ResourceType: "user_group"},
		},
	}
	if flow.GetResourceID() != "res-1" {
		t.Errorf("expected 'res-1', got %q", flow.GetResourceID())
	}

	empty := ApprovalFlow{}
	if empty.GetResourceID() != "" {
		t.Errorf("expected empty string, got %q", empty.GetResourceID())
	}
}
