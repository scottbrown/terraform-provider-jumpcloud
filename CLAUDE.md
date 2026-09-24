# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

The repository follows the HashiCorp terraform-provider-scaffolding-framework layout. CI uses the `GNUmakefile`; a `Taskfile.yml` mirrors it for local convenience.

```bash
make build       # go build ./...
make install     # go install (binary lands in $GOPATH/bin; pair with a dev_overrides block)
make test        # Unit tests
make testacc     # Acceptance tests (TF_ACC=1; requires JUMPCLOUD_API_KEY)
make fmt         # gofmt -s -w -e .
make lint        # golangci-lint run
make generate    # License headers + terraform fmt examples/ + regenerate docs/ (tfplugindocs)

task install     # Build and copy into ~/.terraform.d/plugins/registry.terraform.io/scottbrown/jumpcloud/<version>/
```

Single test: `go test ./internal/resources -run TestUserGroupResource -v`

`docs/` is generated. Edit schemas, `examples/`, or `templates/`, then run `make generate`. CI fails if `docs/` is out of date.

## Environment Variables

- `JUMPCLOUD_API_KEY` — required for the provider to function
- `JUMPCLOUD_ORG_ID` — required for multi-tenant API keys
- `JUMPCLOUD_BASE_URL` — override API base URL (default: `https://console.jumpcloud.com/api/v2`)
- `TF_ACC=1` — enables acceptance tests that make real API calls

## Architecture

Custom Terraform provider using **terraform-plugin-framework** (not the older SDK). All resources use the framework's type system (`types.String`, etc.) and `tfsdk` struct tags.

### Package Layout

- `main.go` — Serves the provider at `registry.terraform.io/scottbrown/jumpcloud`.
- `internal/provider/` — Provider definition, schema, configuration, and registration of all resources and data sources. Acceptance tests (`TestAcc*`) live here and use `testAccProtoV6ProviderFactories`.
- `internal/client/` — HTTP client for the JumpCloud v2 API. One file per API domain. Uses `client.do()` for all requests and `client.IsNotFound()` for 404 handling. Tests use `httptest`.
- `internal/resources/` — Terraform resource implementations (CRUD + import). `attr_helpers.go` has checked accessors for nested-object attributes.
- `internal/datasources/` — Read-only Terraform data source implementations.
- `examples/` — One directory per resource/data source (`resource.tf`, `import.sh`, `data-source.tf`); consumed by tfplugindocs.
- `templates/` — tfplugindocs templates (currently only the provider index page).
- `tools/` — Separate Go module holding `go:generate` directives for copywrite and tfplugindocs.

### Resource Pattern

Each resource follows the same structure:
1. Struct implementing `resource.Resource` (and `ResourceWithImportState`)
2. Model struct with `tfsdk` tags
3. `Configure` extracts `*client.Client` from provider data
4. CRUD methods map between model ↔ API client types
5. `Read` calls `resp.State.RemoveResource(ctx)` on 404 (drift detection)

### Association Resources

`*_association` resources are join-table resources that bind a parent (command, policy, policy group, software app, application) to a target via the JumpCloud associations API (`/v2/{type}/{id}/associations`). Their IDs are composite: `<parent-id>/<target-type>/<target-id>`.

### Provider Config Resolution

Provider attributes (`api_key`, `org_id`, `base_url`) resolve in order: explicit config → environment variable → error/default.

## Linting

`.golangci.yml` is the scaffold's strict config (errcheck, forcetypeassert, godot, unparam, nilerr, ...). Every new Go file needs the license header:

```go
// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MIT
```
