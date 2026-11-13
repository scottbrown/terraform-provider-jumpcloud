// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetPolicy(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/policies/pol123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"id":"pol123","name":"Test Policy","template":{"id":"tmpl1","name":"Custom MDM"},"values":[{"configFieldName":"payload","value":"data"}]}`))
	}))

	policy, err := c.GetPolicy(context.Background(), "pol123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID != "pol123" {
		t.Errorf("expected id 'pol123', got %q", policy.ID)
	}
	if policy.Name != "Test Policy" {
		t.Errorf("expected name 'Test Policy', got %q", policy.Name)
	}
	if policy.TemplateID != "tmpl1" {
		t.Errorf("expected template_id 'tmpl1', got %q", policy.TemplateID)
	}
}

func TestCreatePolicy(t *testing.T) {
	var gotBody map[string]any
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/policies" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"id":"new-pol","name":"Created","template":{"id":"tmpl1"}}`))
	}))

	policy, err := c.CreatePolicy(context.Background(), &Policy{
		Name:       "Created",
		TemplateID: "tmpl1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID != "new-pol" {
		t.Errorf("expected id 'new-pol', got %q", policy.ID)
	}
	if gotBody["template"] != "tmpl1" {
		t.Errorf("expected template 'tmpl1' in request body, got %v", gotBody["template"])
	}
}

func TestDeletePolicy(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/policies/del-pol" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeletePolicy(context.Background(), "del-pol")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPolicy_UnmarshalJSON(t *testing.T) {
	input := `{"id":"p1","name":"Test","template":{"id":"t1","name":"Template Name"},"notes":"hello","values":[{"key":"val"}]}`
	var policy Policy
	err := json.Unmarshal([]byte(input), &policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID != "p1" {
		t.Errorf("expected id 'p1', got %q", policy.ID)
	}
	if policy.TemplateID != "t1" {
		t.Errorf("expected template_id 't1', got %q", policy.TemplateID)
	}
	if policy.Notes != "hello" {
		t.Errorf("expected notes 'hello', got %q", policy.Notes)
	}
}

func TestPolicy_MarshalJSON(t *testing.T) {
	policy := Policy{
		Name:       "Test",
		TemplateID: "tmpl-abc",
		Notes:      "some notes",
		Values:     json.RawMessage(`[{"key":"val"}]`),
	}
	data, err := json.Marshal(&policy)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "Test" {
		t.Errorf("expected name 'Test', got %v", result["name"])
	}
	if result["template"] != "tmpl-abc" {
		t.Errorf("expected template 'tmpl-abc', got %v", result["template"])
	}
}

func TestListPolicies(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		_, _ = w.Write([]byte(`[{"id":"p1","name":"Policy 1","template":{"id":"t1"}},{"id":"p2","name":"Policy 2","template":{"id":"t2"}}]`))
	}))

	policies, err := c.ListPolicies(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(policies))
	}
}
