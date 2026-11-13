schema_version = 1

project {
  license          = "MPL-2.0"
  copyright_holder = "Scott Brown"
  copyright_year   = 2026

  header_ignore = [
    # derived from terraform-provider-scaffolding-framework; these carry the
    # upstream copyright notice alongside this project's and are maintained by hand
    "main.go",
    "internal/provider/provider.go",
    "internal/provider/provider_test.go",
    "tools/tools.go",

    # examples used within documentation (prose)
    "examples/**",

    # GitHub issue template configuration
    ".github/ISSUE_TEMPLATE/*.yml",

    # golangci-lint tooling configuration
    ".golangci.yml",

    # GoReleaser tooling configuration
    ".goreleaser.yml",
  ]
}
