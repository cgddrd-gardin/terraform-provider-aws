package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/directconnect/finder"
)

func TestAccAWSDxConnectionConfirmation_basic(t *testing.T) {
	env, err := testAccCheckAwsDxHostedConnectionEnv()
	if err != nil {
		TestAccSkip(t, err.Error())
	}

	var providers []*schema.Provider

	connectionName := fmt.Sprintf("tf-dx-%s", acctest.RandString(5))
	resourceName := "aws_dx_connection_confirmation.test"
	providerFunc := testAccDxConnectionConfirmationProvider(&providers, 0)
	altProviderFunc := testAccDxConnectionConfirmationProvider(&providers, 1)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ErrorCheck:        testAccErrorCheck(t, directconnect.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxHostedConnectionDestroy(altProviderFunc),
		Steps: []resource.TestStep{
			{
				Config: testAccDxConnectionConfirmationConfig(connectionName, env.ConnectionId, env.OwnerAccountId),
				Check:  testAccCheckAwsDxConnectionConfirmationExists(resourceName, providerFunc),
			},
		},
	})
}

func testAccCheckAwsDxConnectionConfirmationExists(name string, providerFunc func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Direct Connect Connection ID is set")
		}

		provider := providerFunc()
		conn := provider.Meta().(*AWSClient).dxconn

		connection, err := finder.ConnectionByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if state := aws.StringValue(connection.ConnectionState); state != directconnect.ConnectionStateAvailable {
			return fmt.Errorf("Direct Connect Connection %s in unexpected state: %s", rs.Primary.ID, state)
		}

		return nil
	}
}

func testAccDxConnectionConfirmationConfig(name, connectionId, ownerAccountId string) string {
	return composeConfig(
		testAccAlternateAccountProviderConfig(),
		fmt.Sprintf(`
resource "aws_dx_hosted_connection" "connection" {
  provider = "awsalternate"

  name             = "%s"
  connection_id    = "%s"
  owner_account_id = "%s"
  bandwidth        = "100Mbps"
  vlan             = 4092
}

resource "aws_dx_connection_confirmation" "test" {
  connection_id = aws_dx_hosted_connection.connection.id
}
`, name, connectionId, ownerAccountId))
}

func testAccDxConnectionConfirmationProvider(providers *[]*schema.Provider, index int) func() *schema.Provider {
	return func() *schema.Provider {
		return (*providers)[index]
	}
}
