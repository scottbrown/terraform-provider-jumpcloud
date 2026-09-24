// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT

package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testClient(t *testing.T, handler http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return New("test-api-key", "test-org-id", server.URL)
}

func TestNew(t *testing.T) {
	c := New("key", "org", "")
	if c.baseURL != defaultBaseURL {
		t.Errorf("expected default base URL %q, got %q", defaultBaseURL, c.baseURL)
	}

	c = New("key", "org", "http://custom")
	if c.baseURL != "http://custom" {
		t.Errorf("expected custom base URL, got %q", c.baseURL)
	}
}

func TestDo_SetsHeaders(t *testing.T) {
	var gotHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := New("my-key", "my-org", server.URL)
	_ = c.do(context.Background(), "GET", "/test", nil, nil)

	if gotHeaders.Get("x-api-key") != "my-key" {
		t.Errorf("expected x-api-key header 'my-key', got %q", gotHeaders.Get("x-api-key"))
	}
	if gotHeaders.Get("x-org-id") != "my-org" {
		t.Errorf("expected x-org-id header 'my-org', got %q", gotHeaders.Get("x-org-id"))
	}
	if gotHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", gotHeaders.Get("Content-Type"))
	}
}

func TestDo_OmitsOrgIDWhenEmpty(t *testing.T) {
	var gotHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := New("key", "", server.URL)
	_ = c.do(context.Background(), "GET", "/test", nil, nil)

	if gotHeaders.Get("x-org-id") != "" {
		t.Errorf("expected no x-org-id header, got %q", gotHeaders.Get("x-org-id"))
	}
}

func TestDo_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := New("key", "org", server.URL)
	err := c.do(context.Background(), "GET", "/missing", nil, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got: %v", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestDo_DeserializesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "abc123", "name": "test"})
	}))
	defer server.Close()

	c := New("key", "org", server.URL)
	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	err := c.do(context.Background(), "GET", "/resource", nil, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "abc123" {
		t.Errorf("expected id 'abc123', got %q", result.ID)
	}
	if result.Name != "test" {
		t.Errorf("expected name 'test', got %q", result.Name)
	}
}

func TestBuildPaginatedPath(t *testing.T) {
	c := New("key", "org", "")

	tests := []struct {
		name     string
		base     string
		params   *PaginationParams
		expected string
	}{
		{"nil params", "/items", nil, "/items"},
		{"empty params", "/items", &PaginationParams{}, "/items"},
		{"limit only", "/items", &PaginationParams{Limit: 50}, "/items?limit=50"},
		{"skip only", "/items", &PaginationParams{Skip: 10}, "/items?skip=10"},
		{"both", "/items", &PaginationParams{Limit: 100, Skip: 200}, "/items?limit=100&skip=200"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.buildPaginatedPath(tt.base, tt.params)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	if IsNotFound(nil) {
		t.Error("nil should not be not-found")
	}
	if IsNotFound(&APIError{StatusCode: 500}) {
		t.Error("500 should not be not-found")
	}
	if !IsNotFound(&APIError{StatusCode: 404}) {
		t.Error("404 should be not-found")
	}
}
