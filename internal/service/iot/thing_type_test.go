package iot_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_iot_thing_type", &resource.Sweeper{
		Name:         "aws_iot_thing_type",
		F:            sweepThingTypes,
		Dependencies: []string{"aws_iot_thing"},
	})
}

func sweepThingTypes(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListThingTypesInput{}

	err = conn.ListThingTypesPages(input, func(page *iot.ListThingTypesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, thingTypes := range page.ThingTypes {
			r := tfiot.ResourceThingType()
			d := r.Data(nil)

			d.SetId(aws.StringValue(thingTypes.ThingTypeName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Thing Type for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Thing Type for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Thing Type sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSIotThingType_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingTypeConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_thing_type.foo", "arn"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "name", fmt.Sprintf("tf_acc_iot_thing_type_%d", rInt)),
				),
			},
			{
				ResourceName:      "aws_iot_thing_type.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSIotThingType_full(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccThingTypeConfig_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_thing_type.foo", "arn"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "properties.0.description", "MyDescription"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "properties.0.searchable_attributes.#", "3"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "deprecated", "true"),
				),
			},
			{
				ResourceName:      "aws_iot_thing_type.foo",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThingTypeConfig_fullUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "deprecated", "false"),
				),
			},
		},
	})
}

func testAccCheckThingTypeDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_type" {
			continue
		}

		params := &iot.DescribeThingTypeInput{
			ThingTypeName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeThingType(params)
		if err == nil {
			return fmt.Errorf("Expected IoT Thing Type to be destroyed, %s found", rs.Primary.ID)
		}

	}

	return nil
}

func testAccThingTypeConfig_basic(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name = "tf_acc_iot_thing_type_%d"
}
`, rName)
}

func testAccThingTypeConfig_full(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name       = "tf_acc_iot_thing_type_%d"
  deprecated = true

  properties {
    description           = "MyDescription"
    searchable_attributes = ["foo", "bar", "baz"]
  }
}
`, rName)
}

func testAccThingTypeConfig_fullUpdated(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name       = "tf_acc_iot_thing_type_%d"
  deprecated = false

  properties {
    description           = "MyDescription"
    searchable_attributes = ["foo", "bar", "baz"]
  }
}
`, rName)
}
