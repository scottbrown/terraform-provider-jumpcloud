# Terraform Provider for JumpCloud

A [Terraform](https://www.terraform.io) provider for managing [JumpCloud](https://jumpcloud.com) resources through the JumpCloud v2 API. Built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) and laid out per HashiCorp's [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework).

Generated reference documentation lives in [`docs/`](./docs).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (to build the provider)
- A JumpCloud API key

## Using the Provider

```terraform
terraform {
  required_providers {
    jumpcloud = {
      source = "scottbrown/jumpcloud"
    }
  }
}

provider "jumpcloud" {
  # api_key  = var.jumpcloud_api_key   # or JUMPCLOUD_API_KEY
  # org_id   = var.jumpcloud_org_id    # or JUMPCLOUD_ORG_ID (multi-tenant keys only)
  # base_url = "https://console.jumpcloud.com/api/v2"  # or JUMPCLOUD_BASE_URL
}
```

### Authentication

| Attribute  | Environment variable  | Required | Notes                                                      |
|------------|-----------------------|----------|------------------------------------------------------------|
| `api_key`  | `JUMPCLOUD_API_KEY`   | yes      | JumpCloud API key. Marked sensitive.                       |
| `org_id`   | `JUMPCLOUD_ORG_ID`    | no       | Required when the API key is a multi-tenant (MSP) key.     |
| `base_url` | `JUMPCLOUD_BASE_URL`  | no       | Defaults to `https://console.jumpcloud.com/api/v2`.        |

Explicit provider configuration takes precedence over environment variables.

### Resources

| Resource                                | Purpose                                                       |
|-----------------------------------------|---------------------------------------------------------------|
| `jumpcloud_user_group`                  | User groups                                                   |
| `jumpcloud_user_group_membership`       | Bind a user to a user group                                   |
| `jumpcloud_system_group`                | Device (system) groups                                        |
| `jumpcloud_system_group_membership`     | Bind a system to a system group                               |
| `jumpcloud_command`                     | Remote commands                                               |
| `jumpcloud_command_association`         | Target a command at a system or system group                  |
| `jumpcloud_policy`                      | Device policies                                               |
| `jumpcloud_policy_association`          | Target a policy at a system or system group                   |
| `jumpcloud_policy_group`                | Policy groups                                                 |
| `jumpcloud_policy_group_membership`     | Bind a policy to a policy group                               |
| `jumpcloud_policy_group_association`    | Target a policy group at a system or system group             |
| `jumpcloud_password_policy`             | Password complexity and rotation rules                        |
| `jumpcloud_software_app`                | Managed software apps                                         |
| `jumpcloud_software_app_association`    | Target a software app at a system or system group             |
| `jumpcloud_application`                 | SSO applications (SAML / OIDC) with optional SCIM provisioning |
| `jumpcloud_application_association`     | Grant a user or user group access to an application           |
| `jumpcloud_authentication_policy`       | Conditional access policies                                   |
| `jumpcloud_ip_list`                     | IP allow/deny lists used by authentication policies           |
| `jumpcloud_role`                        | Custom admin roles                                            |
| `jumpcloud_approval_flow`               | Approval workflows                                            |
| `jumpcloud_approval_flow_settings`      | Org-wide approval flow settings (singleton, not importable)   |
| `jumpcloud_notification_channel`        | Notification channels                                         |
| `jumpcloud_scheduled_report`            | Scheduled reports                                             |

### Data Sources

`jumpcloud_user_group`, `jumpcloud_system_group`, `jumpcloud_command`, `jumpcloud_policy`, `jumpcloud_policy_template`, `jumpcloud_policy_group`, `jumpcloud_password_policy`, `jumpcloud_software_app`, `jumpcloud_application`, `jumpcloud_authentication_policy`, `jumpcloud_ip_list`, `jumpcloud_role`, `jumpcloud_approval_flow`, `jumpcloud_notification_channel`, `jumpcloud_scheduled_report`.

### Importing

Top-level resources import by JumpCloud ID. Join resources use composite IDs:

```shell
terraform import jumpcloud_user_group.example <group-id>
terraform import jumpcloud_user_group_membership.example <group-id>/<user-id>
terraform import jumpcloud_system_group_membership.example <group-id>/<system-id>
terraform import jumpcloud_policy_group_membership.example <group-id>/<policy-id>
terraform import jumpcloud_command_association.example <command-id>/<target-type>/<target-id>
terraform import jumpcloud_policy_association.example <policy-id>/<target-type>/<target-id>
terraform import jumpcloud_application_association.example <application-id>/<target-type>/<target-id>
```

Each resource page in `docs/resources/` has the exact form.

## Developing the Provider

You need [Go](http://www.golang.org) installed (see [Requirements](#requirements)). The `GNUmakefile` is the canonical entry point and is what CI runs. A `Taskfile.yml` ([go-task](https://taskfile.dev)) mirrors it for convenience.

| Command          | What it does                                                                 |
|------------------|------------------------------------------------------------------------------|
| `make build`     | `go build ./...`                                                             |
| `make install`   | `go install` into `$GOPATH/bin`                                              |
| `make test`      | Unit tests (`httptest`-backed client tests, resource/data-source mapping tests) |
| `make testacc`   | Acceptance tests against a real JumpCloud org (`TF_ACC=1`)                   |
| `make lint`      | `golangci-lint run`                                                          |
| `make fmt`       | `gofmt -s -w -e .`                                                           |
| `make generate`  | Adds license headers, formats `examples/`, regenerates `docs/` with tfplugindocs |
| `task install`   | Builds and copies the binary into `~/.terraform.d/plugins/registry.terraform.io/scottbrown/jumpcloud/<version>/` |

### Running the provider locally

Either use `task install` and reference `scottbrown/jumpcloud` as the source, or point Terraform at a `go install`ed binary with a `dev_overrides` block in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "scottbrown/jumpcloud" = "/path/to/your/GOPATH/bin"
  }
  direct {}
}
```

### Acceptance tests

Acceptance tests create real resources in the JumpCloud org the API key belongs to.

```shell
export JUMPCLOUD_API_KEY=...
export JUMPCLOUD_ORG_ID=...   # multi-tenant keys only
make testacc
```

### Documentation

`docs/` is generated. Edit resource schemas, files under `examples/`, or `templates/`, then run `make generate` and commit the result. CI fails if the generated output is stale.

## License

[MPL-2.0](./LICENSE)
