// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetPasswordPolicy(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/passwordpolicies/pol123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(passwordPolicyGetResponse{
			ObjectID: "pol123",
			Policy: PasswordPolicy{
				Name:            "Strong Password",
				MinLength:       12,
				EnableMinLength: true,
				NeedsUppercase:  true,
				NeedsNumeric:    true,
			},
		})
	}))

	policy, err := c.GetPasswordPolicy(context.Background(), "pol123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID != "pol123" {
		t.Errorf("expected id 'pol123', got %q", policy.ID)
	}
	if policy.Name != "Strong Password" {
		t.Errorf("expected name 'Strong Password', got %q", policy.Name)
	}
	if policy.MinLength != 12 {
		t.Errorf("expected min_length 12, got %d", policy.MinLength)
	}
	if !policy.NeedsUppercase {
		t.Error("expected needs_uppercase to be true")
	}
}

func TestCreatePasswordPolicy(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/passwordpolicies" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body PasswordPolicy
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(passwordPolicyGetResponse{
			ObjectID: "new-id",
			Policy:   body,
		})
	}))

	policy, err := c.CreatePasswordPolicy(context.Background(), &PasswordPolicy{
		Name:           "New Policy",
		MinLength:      8,
		NeedsUppercase: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID != "new-id" {
		t.Errorf("expected id 'new-id', got %q", policy.ID)
	}
	if policy.Name != "New Policy" {
		t.Errorf("expected name 'New Policy', got %q", policy.Name)
	}
}

func TestDeletePasswordPolicy(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/passwordpolicies/del-id" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeletePasswordPolicy(context.Background(), "del-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListPasswordPolicies(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "10" {
			t.Errorf("expected limit=10, got %q", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode(passwordPolicyListResponse{
			Results: []struct {
				ObjectID    string `json:"objectId"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Precedence  int    `json:"precedence"`
				Default     bool   `json:"default"`
			}{
				{ObjectID: "1", Name: "Policy A", Precedence: 1},
				{ObjectID: "2", Name: "Policy B", Precedence: 2},
			},
		})
	}))

	policies, err := c.ListPasswordPolicies(context.Background(), &PaginationParams{Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(policies) != 2 {
		t.Errorf("expected 2 policies, got %d", len(policies))
	}
	if policies[0].ID != "1" {
		t.Errorf("expected first policy id '1', got %q", policies[0].ID)
	}
}

func TestGetPasswordPolicyByName(t *testing.T) {
	callCount := 0
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch r.URL.Path {
		case "/passwordpolicies":
			_ = json.NewEncoder(w).Encode(passwordPolicyListResponse{
				Results: []struct {
					ObjectID    string `json:"objectId"`
					Name        string `json:"name"`
					Description string `json:"description"`
					Precedence  int    `json:"precedence"`
					Default     bool   `json:"default"`
				}{
					{ObjectID: "1", Name: "Weak"},
					{ObjectID: "2", Name: "Strong"},
				},
			})
		case "/passwordpolicies/2":
			_ = json.NewEncoder(w).Encode(passwordPolicyGetResponse{
				ObjectID: "2",
				Policy: PasswordPolicy{
					Name:      "Strong",
					MinLength: 12,
				},
			})
		}
	}))

	policy, err := c.GetPasswordPolicyByName(context.Background(), "Strong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.ID != "2" {
		t.Errorf("expected id '2', got %q", policy.ID)
	}
	if policy.MinLength != 12 {
		t.Errorf("expected min_length 12, got %d", policy.MinLength)
	}
}

func TestGetPasswordPolicyByName_NotFound(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(passwordPolicyListResponse{})
	}))

	_, err := c.GetPasswordPolicyByName(context.Background(), "Missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
