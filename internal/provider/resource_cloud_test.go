package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_ResourceCloud_Basic(t *testing.T) {
	cloudName := acctest.RandomWithPrefix("tfcloud")
	cloudType := "lxd"
	cloudEndpoint := "https://lxd.internal:8443"
	cloudAuthTypes := []string{"certificate"}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceCloudBasic(cloudName, cloudType, cloudEndpoint, cloudAuthTypes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("juju_cloud.this", "name", cloudName),
					resource.TestCheckResourceAttr("juju_cloud.this", "type", cloudType),
					resource.TestCheckResourceAttr("juju_cloud.this", "endpoint", cloudEndpoint),
					resource.TestCheckResourceAttr("juju_cloud.this", "auth_types.#", "1"),
					resource.TestCheckResourceAttr("juju_cloud.this", "auth_types.0", cloudAuthTypes[0]),
				),
			},
			{
				ImportStateVerify: true,
				ImportState:       true,
				ResourceName:      "juju_cloud.this",
			},
		},
	})
}

func testAccResourceCloudBasic(cloudName string, cloudType string, cloudEndpoint string, cloudAuthTypes []string) string {
	return fmt.Sprintf(`
resource "juju_cloud" "this" {
  name = %q
  type = %q
  endpoint = %q
  auth_types = %q
}`, cloudName, cloudType, cloudEndpoint, cloudAuthTypes)
}

func TestAcc_ResourceCloud_Updates(t *testing.T) {
	cloudName := acctest.RandomWithPrefix("tfcloud")
	cloudType := "lxd"
	cloudEndpoint := "https://lxd.internal:8443"
	cloudAuthTypes := []string{"certificate"}
	newCloudEndpoint := "https://my.new.endpoint"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				// Start with the same attributes as above
				Config: testAccResourceCloudUpdates(cloudName, cloudType, cloudEndpoint, cloudAuthTypes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("juju_cloud.this", "name", cloudName),
					resource.TestCheckResourceAttr("juju_cloud.this", "type", cloudType),
					resource.TestCheckResourceAttr("juju_cloud.this", "endpoint", cloudEndpoint),
					resource.TestCheckResourceAttr("juju_cloud.this", "auth_types.#", "1"),
					resource.TestCheckResourceAttr("juju_cloud.this", "auth_types.0", cloudAuthTypes[0]),
				),
			},
			{
				// Change the endpoint URL
				Config: testAccResourceCloudUpdates(cloudName, cloudType, newCloudEndpoint, cloudAuthTypes),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("juju_cloud.this", "endpoint", newCloudEndpoint),
				),
			},
			{
				ImportStateVerify: true,
				ImportState:       true,
				ResourceName:      "juju_cloud.this",
			},
		},
	})
}

func testAccResourceCloudUpdates(cloudName string, cloudType string, cloudEndpoint string, cloudAuthTypes []string) string {
	return fmt.Sprintf(`
resource "juju_cloud" "this" {
  name = %q
  type = %q
  endpoint = %q
  auth_types = %q
}`, cloudName, cloudType, cloudEndpoint, cloudAuthTypes)
}
