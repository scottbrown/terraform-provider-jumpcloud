// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetCommand(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/commands/cmd123" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Command{
			ID:      "cmd123",
			Name:    "Test Command",
			Command: "echo hello",
			User:    "root",
		})
	}))

	cmd, err := c.GetCommand(context.Background(), "cmd123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.ID != "cmd123" {
		t.Errorf("expected id 'cmd123', got %q", cmd.ID)
	}
	if cmd.Command != "echo hello" {
		t.Errorf("expected command 'echo hello', got %q", cmd.Command)
	}
}

func TestCreateCommand(t *testing.T) {
	var gotBody Command
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/commands" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		resp := gotBody
		resp.ID = "new-cmd"
		_ = json.NewEncoder(w).Encode(resp)
	}))

	cmd, err := c.CreateCommand(context.Background(), &Command{
		Name:    "New Command",
		Command: "ls -la",
		User:    "000000000000000000000000",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.ID != "new-cmd" {
		t.Errorf("expected id 'new-cmd', got %q", cmd.ID)
	}
	if gotBody.Name != "New Command" {
		t.Errorf("expected name 'New Command', got %q", gotBody.Name)
	}
}

func TestDeleteCommand(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/commands/del-cmd" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	err := c.DeleteCommand(context.Background(), "del-cmd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListCommands(t *testing.T) {
	c := testClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "50" {
			t.Errorf("expected limit=50, got %q", r.URL.Query().Get("limit"))
		}
		_ = json.NewEncoder(w).Encode(commandListResponse{
			TotalCount: 2,
			Results: []Command{
				{ID: "1", Name: "Cmd A"},
				{ID: "2", Name: "Cmd B"},
			},
		})
	}))

	cmds, err := c.ListCommands(context.Background(), &PaginationParams{Limit: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmds) != 2 {
		t.Errorf("expected 2 commands, got %d", len(cmds))
	}
}
