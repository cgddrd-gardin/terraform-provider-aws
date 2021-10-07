package config_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconfig "github.com/hashicorp/terraform-provider-aws/internal/service/config"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_config_configuration_aggregator", &resource.Sweeper{
		Name: "aws_config_configuration_aggregator",
		F:    sweepConfigurationAggregators,
	})
}

func sweepConfigurationAggregators(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).ConfigConn

	resp, err := conn.DescribeConfigurationAggregators(&configservice.DescribeConfigurationAggregatorsInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Config Configuration Aggregators sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving config configuration aggregators: %s", err)
	}

	if len(resp.ConfigurationAggregators) == 0 {
		log.Print("[DEBUG] No config configuration aggregators to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d config configuration aggregators", len(resp.ConfigurationAggregators))

	for _, agg := range resp.ConfigurationAggregators {
		log.Printf("[INFO] Deleting config configuration aggregator %s", *agg.ConfigurationAggregatorName)
		_, err := conn.DeleteConfigurationAggregator(&configservice.DeleteConfigurationAggregatorInput{
			ConfigurationAggregatorName: agg.ConfigurationAggregatorName,
		})

		if err != nil {
			return fmt.Errorf("error deleting config configuration aggregator %s: %w",
				aws.StringValue(agg.ConfigurationAggregatorName), err)
		}
	}

	return nil
}

func TestAccAWSConfigConfigurationAggregator_account(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	//Name is upper case on purpose to test https://github.com/hashicorp/terraform-provider-aws/issues/8432
	rName := sdkacctest.RandomWithPrefix("Tf-acc-test")
	resourceName := "aws_config_configuration_aggregator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "config", regexp.MustCompile(`config-aggregator/config-aggregator-.+`)),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.0.account_ids.#", "1"),
					acctest.CheckResourceAttrAccountID(resourceName, "account_aggregation_source.0.account_ids.0"),
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.0.regions.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "account_aggregation_source.0.regions.0", "data.aws_region.current", "name"),
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

func TestAccAWSConfigConfigurationAggregator_organization(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_configuration_aggregator.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "organization_aggregation_source.0.role_arn", "aws_iam_role.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.0.all_regions", "true"),
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

func TestAccAWSConfigConfigurationAggregator_switch(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_configuration_aggregator.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "0"),
				),
			},
			{
				Config: testAccConfigurationAggregatorConfig_organization(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "account_aggregation_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "organization_aggregation_source.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSConfigConfigurationAggregator_tags(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_configuration_aggregator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccConfigurationAggregatorTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccConfigurationAggregatorTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					testAccCheckConfigurationAggregatorName(resourceName, rName, &ca),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSConfigConfigurationAggregator_disappears(t *testing.T) {
	var ca configservice.ConfigurationAggregator
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_config_configuration_aggregator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, configservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckConfigurationAggregatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccConfigurationAggregatorConfig_account(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConfigurationAggregatorExists(resourceName, &ca),
					acctest.CheckResourceDisappears(acctest.Provider, tfconfig.ResourceConfigurationAggregator(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckConfigurationAggregatorName(n, desired string, obj *configservice.ConfigurationAggregator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.Attributes["name"] != aws.StringValue(obj.ConfigurationAggregatorName) {
			return fmt.Errorf("expected name: %q, given: %q", desired, aws.StringValue(obj.ConfigurationAggregatorName))
		}
		return nil
	}
}

func testAccCheckConfigurationAggregatorExists(n string, obj *configservice.ConfigurationAggregator) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No config configuration aggregator ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigConn
		out, err := conn.DescribeConfigurationAggregators(&configservice.DescribeConfigurationAggregatorsInput{
			ConfigurationAggregatorNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})
		if err != nil {
			return fmt.Errorf("Failed to describe config configuration aggregator: %s", err)
		}
		if len(out.ConfigurationAggregators) < 1 {
			return fmt.Errorf("No config configuration aggregator found when describing %q", rs.Primary.Attributes["name"])
		}

		ca := out.ConfigurationAggregators[0]
		*obj = *ca

		return nil
	}
}

func testAccCheckConfigurationAggregatorDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ConfigConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_config_configuration_aggregator" {
			continue
		}

		resp, err := conn.DescribeConfigurationAggregators(&configservice.DescribeConfigurationAggregatorsInput{
			ConfigurationAggregatorNames: []*string{aws.String(rs.Primary.Attributes["name"])},
		})

		if err == nil {
			if len(resp.ConfigurationAggregators) != 0 &&
				aws.StringValue(resp.ConfigurationAggregators[0].ConfigurationAggregatorName) == rs.Primary.Attributes["name"] {
				return fmt.Errorf("config configuration aggregator still exists: %s", rs.Primary.Attributes["name"])
			}
		}
	}

	return nil
}

func testAccConfigurationAggregatorConfig_account(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "test" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
  }
}
`, rName)
}

func testAccConfigurationAggregatorConfig_organization(rName string) string {
	return fmt.Sprintf(`
resource "aws_organizations_organization" "test" {
  aws_service_access_principals = ["config.${data.aws_partition.current.dns_suffix}"]
}

resource "aws_config_configuration_aggregator" "test" {
  depends_on = [aws_iam_role_policy_attachment.test, aws_organizations_organization.test]

  name = %[1]q

  organization_aggregation_source {
    all_regions = true
    role_arn    = aws_iam_role.test.arn
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "config.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSConfigRoleForOrganizations"
}
`, rName)
}

func testAccConfigurationAggregatorTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "test" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccConfigurationAggregatorTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_config_configuration_aggregator" "test" {
  name = %[1]q

  account_aggregation_source {
    account_ids = [data.aws_caller_identity.current.account_id]
    regions     = [data.aws_region.current.name]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
