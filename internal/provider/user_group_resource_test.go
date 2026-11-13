// Copyright (c) 2026 Scott Brown
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccUserGroupResource exercises the full lifecycle of a user group and
// reads it back through the data source. It creates a real group in JumpCloud
// and requires TF_ACC=1 plus JUMPCLOUD_API_KEY (and JUMPCLOUD_ORG_ID for
// multi-tenant keys).
func TestAccUserGroupResource(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserGroupResourceConfig(name, "created by acceptance test"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"jumpcloud_user_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
					statecheck.ExpectKnownValue(
						"jumpcloud_user_group.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("created by acceptance test"),
					),
					statecheck.ExpectKnownValue(
						"jumpcloud_user_group.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.jumpcloud_user_group.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(name),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "jumpcloud_user_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccUserGroupResourceConfig(name, "updated by acceptance test"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"jumpcloud_user_group.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("updated by acceptance test"),
					),
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUserGroupResourceConfig(name, description string) string {
	return fmt.Sprintf(`
resource "jumpcloud_user_group" "test" {
  name              = %[1]q
  description       = %[2]q
  membership_method = "STATIC"
}

data "jumpcloud_user_group" "test" {
  id = jumpcloud_user_group.test.id
}
`, name, description)
}
