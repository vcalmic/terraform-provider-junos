package junos_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccJunosSystemLoginUser_basic(t *testing.T) {
	if os.Getenv("TESTACC_SWITCH") == "" {
		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				{
					Config: testAccJunosSystemLoginUserConfigCreate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("junos_system_login_user.testacc",
							"name", "testacc"),
						resource.TestCheckResourceAttrSet("junos_system_login_user.testacc",
							"uid"),
						resource.TestCheckResourceAttr("junos_system_login_user.testacc",
							"authentication.#", "1"),
					),
				},
				{
					Config: testAccJunosSystemLoginUserConfigUpdate(),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("junos_system_login_user.testacc",
							"name", "testacc"),
						resource.TestCheckResourceAttrSet("junos_system_login_user.testacc",
							"uid"),
						resource.TestCheckResourceAttr("junos_system_login_user.testacc",
							"authentication.#", "1"),
						resource.TestCheckResourceAttr("junos_system_login_user.testacc",
							"authentication.0.ssh_public_keys.#", "1"),
					),
				},
				{
					ResourceName:      "junos_system_login_user.testacc",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	}
}

func testAccJunosSystemLoginUserConfigCreate() string {
	return `
resource "junos_system_login_user" "testacc" {
  name       = "testacc"
  class      = "unauthorized"
  cli_prompt = "test cli"
  full_name  = "test name"
  authentication {
    encrypted_password = "test"
    no_public_keys     = true
  }
}
resource "junos_system_login_user" "testacc2" {
  name  = "test.acc2"
  class = "unauthorized"
}
`
}

func testAccJunosSystemLoginUserConfigUpdate() string {
	return `
resource "junos_system_login_user" "testacc" {
  name  = "testacc"
  class = "unauthorized"
  authentication {
    encrypted_password = "test"
    ssh_public_keys    = ["ssh-rsa testkey"]
  }
}
`
}
