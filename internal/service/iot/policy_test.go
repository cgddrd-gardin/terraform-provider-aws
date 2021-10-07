package iot_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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
	resource.AddTestSweepers("aws_iot_policy", &resource.Sweeper{
		Name: "aws_iot_policy",
		F:    sweepPolicies,
		Dependencies: []string{
			"aws_iot_policy_attachment",
		},
	})
}

func sweepPolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListPoliciesInput{}

	err = conn.ListPoliciesPages(input, func(page *iot.ListPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, policy := range page.Policies {
			r := tfiot.ResourcePolicy()
			d := r.Data(nil)

			d.SetId(aws.StringValue(policy.PolicyName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Policy for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Policy for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Policy sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSIoTPolicy_basic(t *testing.T) {
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyInitialStateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "iot", fmt.Sprintf("policy/%s", rName)),
					resource.TestCheckResourceAttrSet(resourceName, "default_version_id"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
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

func TestAccAWSIoTPolicy_disappears(t *testing.T) {
	var v iot.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iot_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyInitialStateConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfiot.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckPolicyDestroy_basic(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_policy" {
			continue
		}

		// Try to find the Policy
		GetPolicyOpts := &iot.GetPolicyInput{
			PolicyName: aws.String(rs.Primary.Attributes["name"]),
		}

		resp, err := conn.GetPolicy(GetPolicyOpts)

		if err == nil {
			if resp.PolicyName != nil {
				return fmt.Errorf("IoT Policy still exists")
			}
		}

		// Verify the error is what we want
		if err != nil {
			iotErr, ok := err.(awserr.Error)
			if !ok || iotErr.Code() != "ResourceNotFoundException" {
				return err
			}
		}
	}

	return nil
}

func testAccCheckPolicyExists(n string, v *iot.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

		resp, err := conn.GetPolicy(&iot.GetPolicyInput{
			PolicyName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPolicyInitialStateConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "test" {
  name = "%s"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iot:*"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF

}
`, rName)
}
