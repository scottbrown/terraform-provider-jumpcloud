// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/scottbrown/terraform-provider-jumpcloud/internal/client"
)

func TestCommandResource_Metadata(t *testing.T) {
	r := NewCommandResource()
	req := resource.MetadataRequest{ProviderTypeName: "jumpcloud"}
	var resp resource.MetadataResponse
	r.Metadata(context.Background(), req, &resp)

	if resp.TypeName != "jumpcloud_command" {
		t.Errorf("expected type name 'jumpcloud_command', got %q", resp.TypeName)
	}
}

func TestCommandResource_Schema(t *testing.T) {
	r := NewCommandResource()
	req := resource.SchemaRequest{}
	var resp resource.SchemaResponse
	r.Schema(context.Background(), req, &resp)

	attrs := resp.Schema.Attributes
	required := []string{"name", "command", "command_type"}
	for _, name := range required {
		attr, ok := attrs[name]
		if !ok {
			t.Errorf("expected attribute %q to exist", name)
			continue
		}
		if !attr.IsRequired() {
			t.Errorf("expected attribute %q to be required", name)
		}
	}

	optional := []string{"shell", "trigger", "schedule", "schedule_repeat", "timeout", "launch_type"}
	for _, name := range optional {
		attr, ok := attrs[name]
		if !ok {
			t.Errorf("expected attribute %q to exist", name)
			continue
		}
		if !attr.IsOptional() {
			t.Errorf("expected attribute %q to be optional", name)
		}
	}
}

func TestCommandResource_ModelToAPI(t *testing.T) {
	r := &CommandResource{}

	model := &CommandResourceModel{
		Name:           types.StringValue("Deploy Script"),
		Command:        types.StringValue("#!/bin/bash\necho hello"),
		CommandType:    types.StringValue("mac"),
		User:           types.StringValue("root"),
		Sudo:           types.BoolValue(true),
		Shell:          types.StringValue("bash"),
		Trigger:        types.StringValue("manual"),
		Schedule:       types.StringValue("0 * * * *"),
		ScheduleRepeat: types.StringValue("daily"),
		Timeout:        types.StringValue("300"),
		LaunchType:     types.StringValue("trigger"),
	}

	cmd := r.modelToAPI(model)

	if cmd.Name != "Deploy Script" {
		t.Errorf("expected name 'Deploy Script', got %q", cmd.Name)
	}
	if cmd.Command != "#!/bin/bash\necho hello" {
		t.Errorf("expected command content, got %q", cmd.Command)
	}
	if cmd.CommandType != "mac" {
		t.Errorf("expected command_type 'mac', got %q", cmd.CommandType)
	}
	if cmd.User != "root" {
		t.Errorf("expected user 'root', got %q", cmd.User)
	}
	if !cmd.Sudo {
		t.Error("expected sudo to be true")
	}
	if cmd.Shell != "bash" {
		t.Errorf("expected shell 'bash', got %q", cmd.Shell)
	}
	if cmd.Trigger != "manual" {
		t.Errorf("expected trigger 'manual', got %q", cmd.Trigger)
	}
	if cmd.Schedule != "0 * * * *" {
		t.Errorf("expected schedule '0 * * * *', got %q", cmd.Schedule)
	}
	if cmd.Timeout != "300" {
		t.Errorf("expected timeout '300', got %q", cmd.Timeout)
	}
	if cmd.LaunchType != "trigger" {
		t.Errorf("expected launch_type 'trigger', got %q", cmd.LaunchType)
	}
}

func TestCommandResource_APIToModel(t *testing.T) {
	r := &CommandResource{}

	t.Run("full fields", func(t *testing.T) {
		cmd := &client.Command{
			ID:          "cmd-abc",
			Name:        "Health Check",
			Command:     "curl localhost",
			CommandType: "linux",
			User:        "nobody",
			Sudo:        false,
			Shell:       "sh",
			Trigger:     "periodic",
			Schedule:    "*/5 * * * *",
			LaunchType:  "scheduled",
		}

		var model CommandResourceModel
		r.apiToModel(cmd, &model)

		if model.Name.ValueString() != "Health Check" {
			t.Errorf("expected name 'Health Check', got %q", model.Name.ValueString())
		}
		if model.CommandType.ValueString() != "linux" {
			t.Errorf("expected command_type 'linux', got %q", model.CommandType.ValueString())
		}
		if model.Shell.ValueString() != "sh" {
			t.Errorf("expected shell 'sh', got %q", model.Shell.ValueString())
		}
		if model.LaunchType.ValueString() != "scheduled" {
			t.Errorf("expected launch_type 'scheduled', got %q", model.LaunchType.ValueString())
		}
	})

	t.Run("empty optional fields stay null", func(t *testing.T) {
		cmd := &client.Command{
			ID:          "cmd-xyz",
			Name:        "Simple",
			Command:     "ls",
			CommandType: "mac",
			User:        "root",
		}

		var model CommandResourceModel
		r.apiToModel(cmd, &model)

		if !model.Shell.IsNull() {
			t.Errorf("expected shell to be null, got %q", model.Shell.ValueString())
		}
		if !model.Trigger.IsNull() {
			t.Errorf("expected trigger to be null, got %q", model.Trigger.ValueString())
		}
		if !model.Schedule.IsNull() {
			t.Errorf("expected schedule to be null, got %q", model.Schedule.ValueString())
		}
		if !model.Timeout.IsNull() {
			t.Errorf("expected timeout to be null, got %q", model.Timeout.ValueString())
		}
		if !model.LaunchType.IsNull() {
			t.Errorf("expected launch_type to be null, got %q", model.LaunchType.ValueString())
		}
	})
}
