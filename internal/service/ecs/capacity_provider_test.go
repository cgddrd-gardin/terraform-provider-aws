package ecs_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_ecs_capacity_provider", &resource.Sweeper{
		Name: "aws_ecs_capacity_provider",
		F:    sweepCapacityProviders,
		Dependencies: []string{
			"aws_ecs_cluster",
			"aws_ecs_service",
		},
	})
}

func sweepCapacityProviders(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ECSConn
	input := &ecs.DescribeCapacityProvidersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = tfecs.DescribeCapacityProvidersPages(conn, input, func(page *ecs.DescribeCapacityProvidersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, capacityProvider := range page.CapacityProviders {
			arn := aws.StringValue(capacityProvider.CapacityProviderArn)

			if name := aws.StringValue(capacityProvider.Name); name == "FARGATE" || name == "FARGATE_SPOT" {
				log.Printf("[INFO] Skipping AWS managed ECS Capacity Provider: %s", arn)
				continue
			}

			r := tfecs.ResourceCapacityProvider()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ECS Capacity Providers sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing ECS Capacity Providers for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping ECS Capacity Providers for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSEcsCapacityProvider_basic(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "id", "ecs", fmt.Sprintf("capacity-provider/%s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "id", resourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10000"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_disappears(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(resourceName, &provider),
					acctest.CheckResourceDisappears(acctest.Provider, tfecs.ResourceCapacityProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_ManagedScaling(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderManagedScalingConfig(rName, ecs.ManagedScalingStatusEnabled, 300, 10, 1, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "1"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "50"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityProviderManagedScalingConfig(rName, ecs.ManagedScalingStatusDisabled, 400, 100, 10, 100),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "400"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
				),
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_ManagedScalingPartial(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderManagedScalingPartialConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "auto_scaling_group_provider.0.auto_scaling_group_arn", "aws_autoscaling_group.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_termination_protection", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.instance_warmup_period", "300"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.minimum_scaling_step_size", "2"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.maximum_scaling_step_size", "10000"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.status", "ENABLED"),
					resource.TestCheckResourceAttr(resourceName, "auto_scaling_group_provider.0.managed_scaling.0.target_capacity", "100"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEcsCapacityProvider_Tags(t *testing.T) {
	var provider ecs.CapacityProvider
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecs_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCapacityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCapacityProviderTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCapacityProviderTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCapacityProviderExists(resourceName, &provider),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckCapacityProviderDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_capacity_provider" {
			continue
		}

		_, err := tfecs.FindCapacityProviderByARN(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("ECS Capacity Provider ID %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCapacityProviderExists(resourceName string, provider *ecs.CapacityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECS Capacity Provider ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn

		output, err := tfecs.FindCapacityProviderByARN(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*provider = *output

		return nil
	}
}

func testAccCapacityProviderBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "test" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_launch_template" "test" {
  image_id      = data.aws_ami.test.id
  instance_type = "t3.micro"
  name          = %[1]q
}

resource "aws_autoscaling_group" "test" {
  availability_zones = data.aws_availability_zones.available.names
  desired_capacity   = 0
  max_size           = 0
  min_size           = 0
  name               = %[1]q

  launch_template {
    id = aws_launch_template.test.id
  }

  tags = [
    {
      key                 = "foo"
      value               = "bar"
      propagate_at_launch = true
    },
  ]
}
`, rName)
}

func testAccCapacityProviderConfig(rName string) string {
	return testAccCapacityProviderBaseConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName)
}

func testAccCapacityProviderManagedScalingConfig(rName, status string, warmup, max, min, cap int) string {
	return testAccCapacityProviderBaseConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.test.arn

    managed_scaling {
      instance_warmup_period    = %[2]d
      maximum_scaling_step_size = %[3]d
      minimum_scaling_step_size = %[4]d
      status                    = %[5]q
      target_capacity           = %[6]d
    }
  }
}
`, rName, warmup, max, min, status, cap)
}

func testAccCapacityProviderManagedScalingPartialConfig(rName string) string {
	return testAccCapacityProviderBaseConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn

    managed_scaling {
      minimum_scaling_step_size = 2
      status                    = "ENABLED"
    }
  }
}
`, rName)
}

func testAccCapacityProviderTags1Config(rName, tag1Key, tag1Value string) string {
	return testAccCapacityProviderBaseConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccCapacityProviderTags2Config(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return testAccCapacityProviderBaseConfig(rName) + fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
