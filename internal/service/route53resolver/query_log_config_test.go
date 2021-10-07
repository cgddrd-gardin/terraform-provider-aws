package route53resolver_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfroute53resolver "github.com/hashicorp/terraform-provider-aws/internal/service/route53resolver"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_route53_resolver_query_log_config", &resource.Sweeper{
		Name: "aws_route53_resolver_query_log_config",
		F:    sweepQueryLogsConfig,
		Dependencies: []string{
			"aws_route53_resolver_query_log_config_association",
		},
	})
}

func sweepQueryLogsConfig(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).Route53ResolverConn
	var sweeperErrs *multierror.Error

	err = conn.ListResolverQueryLogConfigsPages(&route53resolver.ListResolverQueryLogConfigsInput{}, func(page *route53resolver.ListResolverQueryLogConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queryLogConfig := range page.ResolverQueryLogConfigs {
			id := aws.StringValue(queryLogConfig.Id)

			log.Printf("[INFO] Deleting Route53 Resolver Query Log Config: %s", id)
			r := tfroute53resolver.ResourceQueryLogConfig()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Route53 Resolver Query Log Configs sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 Resolver Query Log Configs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSRoute53ResolverQueryLogConfig_basic(t *testing.T) {
	var v route53resolver.ResolverQueryLogConfig
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_query_log_config.test"
	s3BucketResourceName := "aws_s3_bucket.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverQueryLogConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverQueryLogConfigConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverQueryLogConfigExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "destination_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSRoute53ResolverQueryLogConfig_disappears(t *testing.T) {
	var v route53resolver.ResolverQueryLogConfig
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_query_log_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverQueryLogConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverQueryLogConfigConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverQueryLogConfigExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53resolver.ResourceQueryLogConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute53ResolverQueryLogConfig_tags(t *testing.T) {
	var v route53resolver.ResolverQueryLogConfig
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route53_resolver_query_log_config.test"
	cwLogGroupResourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverQueryLogConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverQueryLogConfigConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverQueryLogConfigExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "destination_arn", cwLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverQueryLogConfigConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverQueryLogConfigExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "destination_arn", cwLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRoute53ResolverQueryLogConfigConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverQueryLogConfigExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "destination_arn", cwLogGroupResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckRoute53ResolverQueryLogConfigDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_query_log_config" {
			continue
		}

		// Try to find the resource
		_, err := tfroute53resolver.FindResolverQueryLogConfigByID(conn, rs.Primary.ID)
		// Verify the error is what we want
		if tfawserr.ErrMessageContains(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver Query Log Config still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRoute53ResolverQueryLogConfigExists(n string, v *route53resolver.ResolverQueryLogConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver Query Log Config ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
		out, err := tfroute53resolver.FindResolverQueryLogConfigByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccRoute53ResolverQueryLogConfigConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_route53_resolver_query_log_config" "test" {
  name            = %[1]q
  destination_arn = aws_s3_bucket.test.arn
}
`, rName)
}

func testAccRoute53ResolverQueryLogConfigConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_query_log_config" "test" {
  name            = %[1]q
  destination_arn = aws_cloudwatch_log_group.test.arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRoute53ResolverQueryLogConfigConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_route53_resolver_query_log_config" "test" {
  name            = %[1]q
  destination_arn = aws_cloudwatch_log_group.test.arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
