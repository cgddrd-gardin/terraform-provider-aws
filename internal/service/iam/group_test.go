package iam_test

import (
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_iam_group", &resource.Sweeper{
		Name: "aws_iam_group",
		F:    sweepGroups,
		Dependencies: []string{
			"aws_iam_user",
		},
	})
}

func sweepGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IAMConn
	input := &iam.ListGroupsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListGroupsPages(input, func(page *iam.ListGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, group := range page.Groups {
			name := aws.StringValue(group.GroupName)

			if name == "Admin" || name == "TerraformAccTests" {
				continue
			}

			log.Printf("[INFO] Deleting IAM Group: %s", name)

			getGroupInput := &iam.GetGroupInput{
				GroupName: group.GroupName,
			}

			getGroupOutput, err := conn.GetGroup(getGroupInput)

			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error reading IAM Group (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if getGroupOutput != nil {
				for _, user := range getGroupOutput.Users {
					username := aws.StringValue(user.UserName)

					log.Printf("[INFO] Removing IAM User (%s) from Group: %s", username, name)

					input := &iam.RemoveUserFromGroupInput{
						UserName:  user.UserName,
						GroupName: group.GroupName,
					}

					_, err := conn.RemoveUserFromGroup(input)

					if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
						continue
					}

					if err != nil {
						sweeperErr := fmt.Errorf("error removing IAM User (%s) from IAM Group (%s): %w", username, name, err)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
						continue
					}
				}
			}

			input := &iam.DeleteGroupInput{
				GroupName: group.GroupName,
			}

			if err := tfiam.DeleteGroupPolicyAttachments(conn, name); err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Group (%s) policy attachments: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if err := tfiam.DeleteGroupPolicies(conn, name); err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Group (%s) policies: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = conn.DeleteGroup(input)

			if tfawserr.ErrCodeEquals(err, iam.ErrCodeNoSuchEntityException) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting IAM Group (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM Group sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving IAM Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccIAMGroup_basic(t *testing.T) {
	var conf iam.GetGroupOutput
	resourceName := "aws_iam_group.test"
	resourceName2 := "aws_iam_group.test2"
	rString := sdkacctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-basic-%s", rString)
	groupName2 := fmt.Sprintf("tf-acc-group-basic-2-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &conf),
					testAccCheckGroupAttributes(&conf, groupName, "/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroup2Config(groupName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName2, &conf),
					testAccCheckGroupAttributes(&conf, groupName2, "/funnypath/"),
				),
			},
		},
	})
}

func TestAccIAMGroup_nameChange(t *testing.T) {
	var conf iam.GetGroupOutput
	resourceName := "aws_iam_group.test"
	rString := sdkacctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-basic-%s", rString)
	groupName2 := fmt.Sprintf("tf-acc-group-basic-2-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &conf),
					testAccCheckGroupAttributes(&conf, groupName, "/"),
				),
			},
			{
				Config: testAccGroupConfig(groupName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(resourceName, &conf),
					testAccCheckGroupAttributes(&conf, groupName2, "/"),
				),
			},
		},
	})
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_group" {
			continue
		}

		// Try to get group
		_, err := conn.GetGroup(&iam.GetGroupInput{
			GroupName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return errors.New("still exist.")
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "NoSuchEntity" {
			return err
		}
	}

	return nil
}

func testAccCheckGroupExists(n string, res *iam.GetGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Group name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		resp, err := conn.GetGroup(&iam.GetGroupInput{
			GroupName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckGroupAttributes(group *iam.GetGroupOutput, name string, path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *group.Group.GroupName != name {
			return fmt.Errorf("Bad name: %s when %s was expected", *group.Group.GroupName, name)
		}

		if *group.Group.Path != path {
			return fmt.Errorf("Bad path: %s when %s was expected", *group.Group.Path, path)
		}

		return nil
	}
}

func testAccGroupConfig(groupName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test" {
  name = "%s"
  path = "/"
}
`, groupName)
}

func testAccGroup2Config(groupName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "test2" {
  name = "%s"
  path = "/funnypath/"
}
`, groupName)
}
