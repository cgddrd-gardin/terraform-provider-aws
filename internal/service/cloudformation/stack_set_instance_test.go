package cloudformation_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_cloudformation_stack_set_instance", &resource.Sweeper{
		Name: "aws_cloudformation_stack_set_instance",
		F:    sweepStackSetInstances,
	})
}

func sweepStackSetInstances(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).CloudFormationConn
	input := &cloudformation.ListStackSetsInput{
		Status: aws.String(cloudformation.StackSetStatusActive),
	}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListStackSetsPages(input, func(page *cloudformation.ListStackSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, summary := range page.Summaries {
			input := &cloudformation.ListStackInstancesInput{
				StackSetName: summary.StackSetName,
			}

			err = conn.ListStackInstancesPages(input, func(page *cloudformation.ListStackInstancesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, summary := range page.Summaries {
					r := tfcloudformation.ResourceStackSetInstance()
					d := r.Data(nil)
					id := tfcloudformation.StackSetInstanceCreateResourceID(
						aws.StringValue(summary.StackSetId),
						aws.StringValue(summary.Account),
						aws.StringValue(summary.Region),
					)
					d.SetId(id)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing CloudFormation StackSet Instances (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFormation StackSet Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation StackSets (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping CloudFormation StackSet Instances (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccCloudFormationStackSetInstance_basic(t *testing.T) {
	var stackInstance1 cloudformation.StackInstance
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	cloudformationStackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "region", acctest.Region()),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "stack_id"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_disappears(t *testing.T) {
	var stackInstance1 cloudformation.StackInstance
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudformation.ResourceStackSetInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_Disappears_stackSet(t *testing.T) {
	var stackInstance1 cloudformation.StackInstance
	var stackSet1 cloudformation.StackSet
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	stackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(stackSetResourceName, &stackSet1),
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudformation.ResourceStackSetInstance(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudformation.ResourceStackSet(), stackSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFormationStackSetInstance_parameterOverrides(t *testing.T) {
	var stackInstance1, stackInstance2, stackInstance3, stackInstance4 cloudformation.StackInstance
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceParameterOverrides1Config(rName, "overridevalue1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
			{
				Config: testAccStackSetInstanceParameterOverrides2Config(rName, "overridevalue1updated", "overridevalue2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance2),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance1, &stackInstance2),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter2", "overridevalue2"),
				),
			},
			{
				Config: testAccStackSetInstanceParameterOverrides1Config(rName, "overridevalue1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance3),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance2, &stackInstance3),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
				),
			},
			{
				Config: testAccStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance4),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance3, &stackInstance4),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
				),
			},
		},
	})
}

// TestAccAWSCloudFrontDistribution_RetainStack verifies retain_stack = true
// This acceptance test performs the following steps:
//  * Trigger a Terraform destroy of the resource, which should only remove the instance from the StackSet
//  * Check it still exists outside Terraform
//  * Destroy for real outside Terraform
func TestAccCloudFormationStackSetInstance_retainStack(t *testing.T) {
	var stack1 cloudformation.Stack
	var stackInstance1, stackInstance2, stackInstance3 cloudformation.StackInstance
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckStackSet(t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudformation.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStackSetInstanceRetainStackConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
			{
				Config: testAccStackSetInstanceRetainStackConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance2),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance1, &stackInstance2),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "false"),
				),
			},
			{
				Config: testAccStackSetInstanceRetainStackConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance3),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance2, &stackInstance3),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "true"),
				),
			},
			{
				Config:  testAccStackSetInstanceRetainStackConfig(rName, true),
				Destroy: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceStackExists(&stackInstance3, &stack1),
					testAccCheckCloudFormationStackDisappears(&stack1),
				),
			},
		},
	})
}

func testAccCheckCloudFormationStackSetInstanceExists(resourceName string, v *cloudformation.StackInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn

		stackSetName, accountID, region, err := tfcloudformation.StackSetInstanceParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		output, err := tfcloudformation.FindStackInstanceByName(conn, stackSetName, accountID, region)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckCloudFormationStackSetInstanceStackExists(stackInstance *cloudformation.StackInstance, v *cloudformation.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn

		output, err := tfcloudformation.FindStackByID(conn, aws.StringValue(stackInstance.StackId))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckStackSetInstanceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFormationConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudformation_stack_set_instance" {
			continue
		}

		stackSetName, accountID, region, err := tfcloudformation.StackSetInstanceParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfcloudformation.FindStackInstanceByName(conn, stackSetName, accountID, region)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("CloudFormation StackSet Instance %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCloudFormationStackSetInstanceNotRecreated(i, j *cloudformation.StackInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StackId) != aws.StringValue(j.StackId) {
			return fmt.Errorf("CloudFormation StackSet Instance (%s,%s,%s) recreated", aws.StringValue(i.StackSetId), aws.StringValue(i.Account), aws.StringValue(i.Region))
		}

		return nil
	}
}

func testAccStackSetInstanceBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "Administration" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Administration"
}

resource "aws_iam_role_policy" "Administration" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Administration.name
}

resource "aws_iam_role" "Execution" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "${aws_iam_role.Administration.arn}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Execution"
}

resource "aws_iam_role_policy" "Execution" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "*"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Execution.name
}

resource "aws_cloudformation_stack_set" "test" {
  depends_on = [aws_iam_role_policy.Execution]

  administration_role_arn = aws_iam_role.Administration.arn
  execution_role_name     = aws_iam_role.Execution.name
  name                    = %[1]q

  parameters = {
    Parameter1 = "stacksetvalue1"
    Parameter2 = "stacksetvalue2"
  }

  template_body = <<TEMPLATE
Parameters:
  Parameter1:
    Type: String
  Parameter2:
    Type: String
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Parameter2Value:
    Value: !Ref Parameter2
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
TEMPLATE
}
`, rName)
}

func testAccStackSetInstanceConfig(rName string) string {
	return testAccStackSetInstanceBaseConfig(rName) + `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`
}

func testAccStackSetInstanceParameterOverrides1Config(rName, value1 string) string {
	return testAccStackSetInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  parameter_overrides = {
    Parameter1 = %[1]q
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, value1)
}

func testAccStackSetInstanceParameterOverrides2Config(rName, value1, value2 string) string {
	return testAccStackSetInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  parameter_overrides = {
    Parameter1 = %[1]q
    Parameter2 = %[2]q
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, value1, value2)
}

func testAccStackSetInstanceRetainStackConfig(rName string, retainStack bool) string {
	return testAccStackSetInstanceBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  retain_stack   = %[1]t
  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, retainStack)
}
