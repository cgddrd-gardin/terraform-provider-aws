package cloudwatchlogs_test

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudwatchlogs "github.com/hashicorp/terraform-provider-aws/internal/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_query_definition", &resource.Sweeper{
		Name: "aws_cloudwatch_query_definition",
		F:    sweeplogQueryDefinitions,
	})
}

func sweeplogQueryDefinitions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).CloudWatchLogsConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &cloudwatchlogs.DescribeQueryDefinitionsInput{}

	// AWS SDK Go does not currently provide paginator
	for {
		output, err := conn.DescribeQueryDefinitions(input)

		if err != nil {
			err := fmt.Errorf("error reading CloudWatch Log Query Definition: %w", err)
			log.Printf("[ERROR] %s", err)
			errs = multierror.Append(errs, err)
			break
		}

		for _, queryDefinition := range output.QueryDefinitions {
			r := tfcloudwatchlogs.ResourceQueryDefinition()
			d := r.Data(nil)

			d.SetId(aws.StringValue(queryDefinition.QueryDefinitionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping CloudWatch Log Query Definition for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping CloudWatch Log Query Definition sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccCloudWatchLogsQueryDefinition_basic(t *testing.T) {
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix("tf-acc-test")

	expectedQueryString := `fields @timestamp, @message
| sort @timestamp desc
| limit 20
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "query_string", expectedQueryString),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "query_definition_id", regexp.MustCompile(verify.UUIDRegexPattern)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccQueryDefinitionImportStateID(&v),
			},
		},
	})
}

func testAccQueryDefinitionImportStateID(v *cloudwatchlogs.QueryDefinition) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		id := arn.ARN{
			AccountID: acctest.AccountID(),
			Partition: acctest.Partition(),
			Region:    acctest.Region(),
			Service:   cloudwatchlogs.ServiceName,
			Resource:  fmt.Sprintf("query-definition:%s", aws.StringValue(v.QueryDefinitionId)),
		}

		return id.String(), nil
	}
}

func TestAccCloudWatchLogsQueryDefinition_disappears(t *testing.T) {
	var v cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudwatchlogs.ResourceQueryDefinition(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudWatchLogsQueryDefinition_rename(t *testing.T) {
	var v1, v2 cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix("tf-acc-test")
	updatedQueryName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_Basic(queryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
				),
			},
			{
				Config: testAccQueryDefinitionConfig_Basic(updatedQueryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", updatedQueryName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccQueryDefinitionImportStateID(&v2),
			},
		},
	})
}

func TestAccCloudWatchLogsQueryDefinition_logGroups(t *testing.T) {
	var v1, v2 cloudwatchlogs.QueryDefinition
	resourceName := "aws_cloudwatch_query_definition.test"
	queryName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, cloudwatchlogs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckQueryDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryDefinitionConfig_LogGroups(queryName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.0", "aws_cloudwatch_log_group.test.0", "name"),
				),
			},
			{
				Config: testAccQueryDefinitionConfig_LogGroups(queryName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckQueryDefinitionExists(resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "name", queryName),
					resource.TestCheckResourceAttr(resourceName, "log_group_names.#", "5"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.0", "aws_cloudwatch_log_group.test.0", "name"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_names.1", "aws_cloudwatch_log_group.test.1", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccQueryDefinitionImportStateID(&v2),
			},
		},
	})
}

func testAccCheckQueryDefinitionExists(rName string, v *cloudwatchlogs.QueryDefinition) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchLogsConn

		result, err := tfcloudwatchlogs.FindQueryDefinition(context.Background(), conn, "", rs.Primary.ID)

		if err != nil {
			return err
		}

		if result == nil {
			return fmt.Errorf("CloudWatch query definition (%s) not found", rs.Primary.ID)
		}

		*v = *result

		return nil
	}
}

func testAccCheckQueryDefinitionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudWatchLogsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_query_definition" {
			continue
		}

		result, err := tfcloudwatchlogs.FindQueryDefinition(context.Background(), conn, "", rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error reading CloudWatch query definition (%s): %w", rs.Primary.ID, err)
		}

		if result != nil {
			return fmt.Errorf("CloudWatch query definition (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccQueryDefinitionConfig_Basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_query_definition" "test" {
  name = %[1]q

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}
`, rName)
}

func testAccQueryDefinitionConfig_LogGroups(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_query_definition" "test" {
  name = %[1]q

  log_group_names = aws_cloudwatch_log_group.test[*].name

  query_string = <<EOF
fields @timestamp, @message
| sort @timestamp desc
| limit 20
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  count = %[2]d

  name = "%[1]s-${count.index}"
}
`, rName, count)
}
