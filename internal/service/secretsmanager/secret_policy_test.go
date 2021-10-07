package secretsmanager_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsecretsmanager "github.com/hashicorp/terraform-provider-aws/internal/service/secretsmanager"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_secretsmanager_secret_policy", &resource.Sweeper{
		Name: "aws_secretsmanager_secret_policy",
		F:    sweepSecretPolicies,
	})
}

func sweepSecretPolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SecretsManagerConn

	err = conn.ListSecretsPages(&secretsmanager.ListSecretsInput{}, func(page *secretsmanager.ListSecretsOutput, lastPage bool) bool {
		if len(page.SecretList) == 0 {
			log.Print("[DEBUG] No Secrets Manager Secrets to sweep")
			return true
		}

		for _, secret := range page.SecretList {
			name := aws.StringValue(secret.Name)

			log.Printf("[INFO] Deleting Secrets Manager Secret Policy: %s", name)
			input := &secretsmanager.DeleteResourcePolicyInput{
				SecretId: aws.String(name),
			}

			_, err := conn.DeleteResourcePolicy(input)
			if err != nil {
				if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
					continue
				}
				log.Printf("[ERROR] Failed to delete Secrets Manager Secret Policy (%s): %s", name, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Secrets Manager Secret sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Secrets Manager Secrets: %w", err)
	}
	return nil
}

func TestAccAwsSecretsManagerSecretPolicy_basic(t *testing.T) {
	var policy secretsmanager.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSecretPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(resourceName, &policy),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(`{"Action":"secretsmanager:GetSecretValue".+`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"block_public_policy"},
			},
			{
				Config: testAccSecretPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(resourceName, &policy),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile(`{"Action":"secretsmanager:\*".+`)),
				),
			},
		},
	})
}

func TestAccAwsSecretsManagerSecretPolicy_blockPublicPolicy(t *testing.T) {
	var policy secretsmanager.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSecretPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretPolicyBlockConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"block_public_policy"},
			},
			{
				Config: testAccSecretPolicyBlockConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "false"),
				),
			},
			{
				Config: testAccSecretPolicyBlockConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(resourceName, &policy),
					resource.TestCheckResourceAttr(resourceName, "block_public_policy", "true"),
				),
			},
		},
	})
}

func TestAccAwsSecretsManagerSecretPolicy_disappears(t *testing.T) {
	var policy secretsmanager.GetResourcePolicyOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_secretsmanager_secret_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, secretsmanager.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSecretPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSecretPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSecretPolicyExists(resourceName, &policy),
					acctest.CheckResourceDisappears(acctest.Provider, tfsecretsmanager.ResourceSecretPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSecretPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_secretsmanager_secret_policy" {
			continue
		}

		secretInput := &secretsmanager.DescribeSecretInput{
			SecretId: aws.String(rs.Primary.ID),
		}

		var output *secretsmanager.DescribeSecretOutput

		err := resource.Retry(tfsecretsmanager.PropagationTimeout, func() *resource.RetryError {
			var err error
			output, err = conn.DescribeSecret(secretInput)

			if err != nil {
				return resource.NonRetryableError(err)
			}

			if output != nil && output.DeletedDate == nil {
				return resource.RetryableError(fmt.Errorf("Secret %q still exists", rs.Primary.ID))
			}

			return nil
		})

		if tfresource.TimedOut(err) {
			output, err = conn.DescribeSecret(secretInput)
		}

		if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.DeletedDate == nil {
			return fmt.Errorf("Secret %q still exists", rs.Primary.ID)
		}

		input := &secretsmanager.GetResourcePolicyInput{
			SecretId: aws.String(rs.Primary.ID),
		}

		_, err = conn.GetResourcePolicy(input)

		if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeResourceNotFoundException, "") ||
			tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeInvalidRequestException,
				"You can't perform this operation on the secret because it was marked for deletion.") {
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil

}

func testAccCheckSecretPolicyExists(resourceName string, policy *secretsmanager.GetResourcePolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecretsManagerConn
		input := &secretsmanager.GetResourcePolicyInput{
			SecretId: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetResourcePolicy(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Secret Policy %q does not exist", rs.Primary.ID)
		}

		*policy = *output

		return nil
	}
}

func testAccSecretPolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn = aws_secretsmanager_secret.test.arn

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Sid": "EnableAllPermissions",
	  "Effect": "Allow",
	  "Principal": {
		"AWS": "${aws_iam_role.test.arn}"
	  },
	  "Action": "secretsmanager:GetSecretValue",
	  "Resource": "*"
	}
  ]
}
POLICY
}
`, rName)
}

func testAccSecretPolicyUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn = aws_secretsmanager_secret.test.arn

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Sid": "EnableAllPermissions",
	  "Effect": "Allow",
	  "Principal": {
		"AWS": "*"
	  },
	  "Action": "secretsmanager:*",
	  "Resource": "*"
	}
  ]
}
POLICY
}
`, rName)
}

func testAccSecretPolicyBlockConfig(rName string, block bool) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_policy" "test" {
  secret_arn          = aws_secretsmanager_secret.test.arn
  block_public_policy = %[2]t

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Sid": "EnableAllPermissions",
	  "Effect": "Allow",
	  "Principal": {
		"AWS": "${aws_iam_role.test.arn}"
	  },
	  "Action": "secretsmanager:GetSecretValue",
	  "Resource": "*"
	}
  ]
}
POLICY
}
`, rName, block)
}
