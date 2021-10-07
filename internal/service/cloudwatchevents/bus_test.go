package cloudwatchevents_test

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchevents "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchevents"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_event_bus", &resource.Sweeper{
		Name: "aws_cloudwatch_event_bus",
		F:    sweepBuses,
		Dependencies: []string{
			"aws_cloudwatch_event_rule",
			"aws_cloudwatch_event_target",
			"aws_schemas_discoverer",
		},
	})
}

func sweepBuses(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudWatchEventsConn
	input := &events.ListEventBusesInput{}
	var sweeperErrs *multierror.Error

	err = tfcloudwatchevents.ListEventBusesPages(conn, input, func(page *events.ListEventBusesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventBus := range page.EventBuses {
			name := aws.StringValue(eventBus.Name)
			if name == tfcloudwatchevents.DefaultEventBusName {
				continue
			}

			r := tfcloudwatchevents.ResourceBus()
			d := r.Data(nil)
			d.SetId(name)
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Events event bus sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudWatch Events event buses: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudWatchEventBus_basic(t *testing.T) {
	var v1, v2, v3 events.DescribeEventBusOutput
	busName := sdkacctest.RandomWithPrefix("tf-acc-test")
	busNameModified := sdkacctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v1),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckNoResourceAttr(resourceName, "event_source_name"),
					resource.TestCheckResourceAttr(resourceName, "name", busName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBusConfig(busNameModified),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v2),
					testAccCheckCloudWatchEventBusRecreated(&v1, &v2),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "events", fmt.Sprintf("event-bus/%s", busNameModified)),
					resource.TestCheckNoResourceAttr(resourceName, "event_source_name"),
					resource.TestCheckResourceAttr(resourceName, "name", busNameModified),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccBusConfig_Tags1(busNameModified, "key", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v3),
					testAccCheckCloudWatchEventBusNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventBus_tags(t *testing.T) {
	var v1, v2, v3, v4 events.DescribeEventBusOutput
	busName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig_Tags1(busName, "key1", "value"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBusConfig_Tags2(busName, "key1", "updated", "key2", "added"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v2),
					testAccCheckCloudWatchEventBusNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccBusConfig_Tags1(busName, "key2", "added"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v3),
					testAccCheckCloudWatchEventBusNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "added"),
				),
			},
			{
				Config: testAccBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v4),
					testAccCheckCloudWatchEventBusNotRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchEventBus_default(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccBusConfig("default"),
				ExpectError: regexp.MustCompile(`cannot be 'default'`),
			},
		},
	})
}

func TestAccAWSCloudWatchEventBus_disappears(t *testing.T) {
	var v events.DescribeEventBusOutput
	busName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatchevents.ResourceBus(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudWatchEventBus_PartnerEventSource(t *testing.T) {
	key := "EVENT_BRIDGE_PARTNER_EVENT_SOURCE_NAME"
	busName := os.Getenv(key)
	if busName == "" {
		t.Skipf("Environment variable %s is not set", key)
	}

	var busOutput events.DescribeEventBusOutput
	resourceName := "aws_cloudwatch_event_bus.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, events.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBusPartnerEventSourceConfig(busName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchEventBusExists(resourceName, &busOutput),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "events", fmt.Sprintf("event-bus/%s", busName)),
					resource.TestCheckResourceAttr(resourceName, "event_source_name", busName),
					resource.TestCheckResourceAttr(resourceName, "name", busName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckBusDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEventsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_event_bus" {
			continue
		}

		params := events.DescribeEventBusInput{
			Name: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeEventBus(&params)

		if err == nil {
			return fmt.Errorf("CloudWatch Events event bus (%s) still exists: %s", rs.Primary.ID, resp)
		}
	}

	return nil
}

func testAccCheckCloudWatchEventBusExists(n string, v *events.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchEventsConn
		params := events.DescribeEventBusInput{
			Name: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeEventBus(&params)
		if err != nil {
			return err
		}
		if resp == nil {
			return fmt.Errorf("CloudWatch Events event bus (%s) not found", n)
		}

		*v = *resp

		return nil
	}
}

func testAccCheckCloudWatchEventBusRecreated(i, j *events.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Arn) == aws.StringValue(j.Arn) {
			return fmt.Errorf("CloudWatch Events event bus not recreated")
		}
		return nil
	}
}

func testAccCheckCloudWatchEventBusNotRecreated(i, j *events.DescribeEventBusOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.Arn) != aws.StringValue(j.Arn) {
			return fmt.Errorf("CloudWatch Events event bus was recreated")
		}
		return nil
	}
}

func testAccBusConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}
`, name)
}

func testAccBusConfig_Tags1(name, key, value string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, name, key, value)
}

func testAccBusConfig_Tags2(name, key1, value1, key2, value2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, key1, value1, key2, value2)
}

func testAccBusPartnerEventSourceConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name              = %[1]q
  event_source_name = %[1]q
}
`, name)
}
