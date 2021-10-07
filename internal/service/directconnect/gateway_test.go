package directconnect_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdirectconnect "github.com/hashicorp/terraform-provider-aws/internal/service/directconnect"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_dx_gateway", &resource.Sweeper{
		Name: "aws_dx_gateway",
		F:    sweepGateways,
		Dependencies: []string{
			"aws_dx_gateway_association",
		},
	})
}

func sweepGateways(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).DirectConnectConn
	input := &directconnect.DescribeDirectConnectGatewaysInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = tfdirectconnect.DescribeDirectConnectGatewaysPages(conn, input, func(page *directconnect.DescribeDirectConnectGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gateway := range page.DirectConnectGateways {
			directConnectGatewayID := aws.StringValue(gateway.DirectConnectGatewayId)

			if state := aws.StringValue(gateway.DirectConnectGatewayState); state != directconnect.GatewayStateAvailable {
				log.Printf("[INFO] Skipping Direct Connect Gateway in non-available (%s) state: %s", state, directConnectGatewayID)
				continue
			}

			input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				DirectConnectGatewayId: aws.String(directConnectGatewayID),
			}

			var associations bool

			err := tfdirectconnect.DescribeDirectConnectGatewayAssociationsPages(conn, input, func(page *directconnect.DescribeDirectConnectGatewayAssociationsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				// If associations still remain, its likely that our region is not the home
				// region of those associations and the previous sweepers skipped them.
				// When we hit this condition, we skip trying to delete the gateway as it
				// will go from deleting -> available after a few minutes and timeout.
				if len(page.DirectConnectGatewayAssociations) > 0 {
					associations = true

					return false
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateway Associations (%s): %w", region, err))
			}

			if associations {
				log.Printf("[INFO] Skipping Direct Connect Gateway with remaining associations: %s", directConnectGatewayID)
				continue
			}

			r := tfdirectconnect.ResourceGateway()
			d := r.Data(nil)
			d.SetId(directConnectGatewayID)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Print(fmt.Errorf("[WARN] Skipping Direct Connect Gateway sweep for %s: %w", region, err))
		return sweeperErrs // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Direct Connect Gateways (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Direct Connect Gateways (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsDxGateway_basic(t *testing.T) {
	var v directconnect.Gateway
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGateway_disappears(t *testing.T) {
	var v directconnect.Gateway
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfdirectconnect.ResourceGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsDxGateway_complex(t *testing.T) {
	var v directconnect.Gateway
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, directconnect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_multiVpnGatewaysSingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(resourceName, &v),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_account_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_gateway" {
			continue
		}

		_, err := tfdirectconnect.FindGatewayByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Direct Connect Gateway %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckGatewayExists(name string, v *directconnect.Gateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DirectConnectConn

		output, err := tfdirectconnect.FindGatewayByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccDxGatewayConfig(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = %[1]q
  amazon_side_asn = "%[2]d"
}
`, rName, rBgpAsn)
}
