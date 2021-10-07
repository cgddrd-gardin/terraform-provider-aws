package iam_test

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
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
	"github.com/pquerna/otp/totp"
)

func init() {
	resource.AddTestSweepers("aws_iam_user", &resource.Sweeper{
		Name: "aws_iam_user",
		F:    sweepUsers,
	})
}

func sweepUsers(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).IAMConn
	prefixes := []string{
		"test-user",
		"test_user",
		"tf-acc",
		"tf_acc",
	}
	users := make([]*iam.User, 0)

	err = conn.ListUsersPages(&iam.ListUsersInput{}, func(page *iam.ListUsersOutput, lastPage bool) bool {
		for _, user := range page.Users {
			for _, prefix := range prefixes {
				if strings.HasPrefix(aws.StringValue(user.UserName), prefix) {
					users = append(users, user)
					break
				}
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM User sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving IAM Users: %s", err)
	}

	if len(users) == 0 {
		log.Print("[DEBUG] No IAM Users to sweep")
		return nil
	}

	var sweeperErrs *multierror.Error
	for _, user := range users {
		username := aws.StringValue(user.UserName)
		log.Printf("[DEBUG] Deleting IAM User: %s", username)

		listUserPoliciesInput := &iam.ListUserPoliciesInput{
			UserName: user.UserName,
		}
		listUserPoliciesOutput, err := conn.ListUserPolicies(listUserPoliciesInput)

		if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error listing IAM User (%s) inline policies: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, inlinePolicyName := range listUserPoliciesOutput.PolicyNames {
			log.Printf("[DEBUG] Deleting IAM User (%s) inline policy %q", username, *inlinePolicyName)

			input := &iam.DeleteUserPolicyInput{
				PolicyName: inlinePolicyName,
				UserName:   user.UserName,
			}

			if _, err := conn.DeleteUserPolicy(input); err != nil {
				if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
					continue
				}
				sweeperErr := fmt.Errorf("error deleting IAM User (%s) inline policy %q: %s", username, *inlinePolicyName, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		listAttachedUserPoliciesInput := &iam.ListAttachedUserPoliciesInput{
			UserName: user.UserName,
		}
		listAttachedUserPoliciesOutput, err := conn.ListAttachedUserPolicies(listAttachedUserPoliciesInput)

		if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error listing IAM User (%s) attached policies: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, attachedPolicy := range listAttachedUserPoliciesOutput.AttachedPolicies {
			policyARN := aws.StringValue(attachedPolicy.PolicyArn)

			log.Printf("[DEBUG] Detaching IAM User (%s) attached policy: %s", username, policyARN)

			if err := tfiam.DetachPolicyFromUser(conn, username, policyARN); err != nil {
				sweeperErr := fmt.Errorf("error detaching IAM User (%s) attached policy (%s): %s", username, policyARN, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if err := tfiam.DeleteUserGroupMemberships(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) group memberships: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := tfiam.DeleteUserAccessKeys(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) access keys: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := tfiam.DeleteUserSSHKeys(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) SSH keys: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := tfiam.DeleteUserVirtualMFADevices(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) virtual MFA devices: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := tfiam.DeactivateUserMFADevices(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) MFA devices: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		if err := tfiam.DeleteUserLoginProfile(conn, username); err != nil {
			sweeperErr := fmt.Errorf("error removing IAM User (%s) login profile: %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		input := &iam.DeleteUserInput{
			UserName: aws.String(username),
		}

		_, err = conn.DeleteUser(input)

		if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error deleting IAM User (%s): %s", username, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSUser_basic(t *testing.T) {
	var conf iam.GetUserOutput

	name1 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	path1 := "/"
	path2 := "/path2/"
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(name1, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("aws_iam_user.user", &conf),
					testAccCheckUserAttributes(&conf, name1, "/"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
			{
				Config: testAccUserConfig(name2, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("aws_iam_user.user", &conf),
					testAccCheckUserAttributes(&conf, name2, "/path2/"),
				),
			},
		},
	})
}

func TestAccAWSUser_disappears(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserDisappears(&user),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSUser_ForceDestroy_AccessKey(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserForceDestroyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserCreatesAccessKey(&user),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccAWSUser_ForceDestroy_LoginProfile(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserForceDestroyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserCreatesLoginProfile(&user),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccAWSUser_ForceDestroy_MFADevice(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserForceDestroyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserCreatesMFADevice(&user),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccAWSUser_ForceDestroy_SSHKey(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserForceDestroyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserUploadsSSHKey(&user),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func TestAccAWSUser_ForceDestroy_SigningCertificate(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserForceDestroyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					testAccCheckUserUploadSigningCertificate(&user),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
		},
	})
}

func TestAccAWSUser_nameChange(t *testing.T) {
	var conf iam.GetUserOutput

	name1 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	name2 := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	path := "/"
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(name1, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("aws_iam_user.user", &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
			{
				Config: testAccUserConfig(name2, path),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("aws_iam_user.user", &conf),
				),
			},
		},
	})
}

func TestAccAWSUser_pathChange(t *testing.T) {
	var conf iam.GetUserOutput

	name := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	path1 := "/"
	path2 := "/updated/"
	resourceName := "aws_iam_user.user"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(name, path1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("aws_iam_user.user", &conf),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
			{
				Config: testAccUserConfig(name, path2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists("aws_iam_user.user", &conf),
				),
			},
		},
	})
}

func TestAccAWSUser_permissionsBoundary(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.user"

	permissionsBoundary1 := fmt.Sprintf("arn:%s:iam::aws:policy/AdministratorAccess", acctest.Partition())
	permissionsBoundary2 := fmt.Sprintf("arn:%s:iam::aws:policy/ReadOnlyAccess", acctest.Partition())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: testAccUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckUserPermissionsBoundary(&user, permissionsBoundary1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
			// Test update
			{
				Config: testAccUserConfig_permissionsBoundary(rName, permissionsBoundary2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary2),
					testAccCheckUserPermissionsBoundary(&user, permissionsBoundary2),
				),
			},
			// Test import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
			// Test removal
			{
				Config: testAccUserConfig(rName, "/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckUserPermissionsBoundary(&user, ""),
				),
			},
			// Test addition
			{
				Config: testAccUserConfig_permissionsBoundary(rName, permissionsBoundary1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", permissionsBoundary1),
					testAccCheckUserPermissionsBoundary(&user, permissionsBoundary1),
				),
			},
			// Test empty value
			{
				Config: testAccUserConfig_permissionsBoundary(rName, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "permissions_boundary", ""),
					testAccCheckUserPermissionsBoundary(&user, ""),
				),
			},
		},
	})
}

func TestAccAWSUser_tags(t *testing.T) {
	var user iam.GetUserOutput

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "test-Name"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-tag2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force_destroy"},
			},
			{
				Config: testAccUserConfig_tagsUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(resourceName, &user),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-tagUpdate"),
				),
			},
		},
	})
}

func testAccCheckUserDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_user" {
			continue
		}

		// Try to get user
		_, err := conn.GetUser(&iam.GetUserInput{
			UserName: aws.String(rs.Primary.ID),
		})
		if err == nil {
			return fmt.Errorf("still exist.")
		}

		// Verify the error is what we want
		if !tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
			return err
		}
	}

	return nil
}

func testAccCheckUserExists(n string, res *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		resp, err := conn.GetUser(&iam.GetUserInput{
			UserName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckUserAttributes(user *iam.GetUserOutput, name string, path string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *user.User.UserName != name {
			return fmt.Errorf("Bad name: %s", *user.User.UserName)
		}

		if *user.User.Path != path {
			return fmt.Errorf("Bad path: %s", *user.User.Path)
		}

		return nil
	}
}

func testAccCheckUserDisappears(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		userName := aws.StringValue(getUserOutput.User.UserName)

		_, err := conn.DeleteUser(&iam.DeleteUserInput{
			UserName: aws.String(userName),
		})
		if err != nil {
			return fmt.Errorf("error deleting user %q: %s", userName, err)
		}

		return nil
	}
}

func testAccCheckUserPermissionsBoundary(getUserOutput *iam.GetUserOutput, expectedPermissionsBoundaryArn string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualPermissionsBoundaryArn := ""

		if getUserOutput.User.PermissionsBoundary != nil {
			actualPermissionsBoundaryArn = *getUserOutput.User.PermissionsBoundary.PermissionsBoundaryArn
		}

		if actualPermissionsBoundaryArn != expectedPermissionsBoundaryArn {
			return fmt.Errorf("PermissionsBoundary: '%q', expected '%q'.", actualPermissionsBoundaryArn, expectedPermissionsBoundaryArn)
		}

		return nil
	}
}

func testAccCheckUserCreatesAccessKey(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		input := &iam.CreateAccessKeyInput{
			UserName: getUserOutput.User.UserName,
		}

		if _, err := conn.CreateAccessKey(input); err != nil {
			return fmt.Errorf("error creating IAM User (%s) Access Key: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserCreatesLoginProfile(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		password, err := tfiam.GeneratePassword(32)
		if err != nil {
			return err
		}
		input := &iam.CreateLoginProfileInput{
			Password: aws.String(password),
			UserName: getUserOutput.User.UserName,
		}

		if _, err := conn.CreateLoginProfile(input); err != nil {
			return fmt.Errorf("error creating IAM User (%s) Login Profile: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserCreatesMFADevice(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		createVirtualMFADeviceInput := &iam.CreateVirtualMFADeviceInput{
			Path:                 getUserOutput.User.Path,
			VirtualMFADeviceName: getUserOutput.User.UserName,
		}

		createVirtualMFADeviceOutput, err := conn.CreateVirtualMFADevice(createVirtualMFADeviceInput)
		if err != nil {
			return fmt.Errorf("error creating IAM User (%s) Virtual MFA Device: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		secret := string(createVirtualMFADeviceOutput.VirtualMFADevice.Base32StringSeed)
		authenticationCode1, err := totp.GenerateCode(secret, time.Now().Add(-30*time.Second))
		if err != nil {
			return fmt.Errorf("error generating Virtual MFA Device authentication code 1: %s", err)
		}
		authenticationCode2, err := totp.GenerateCode(secret, time.Now())
		if err != nil {
			return fmt.Errorf("error generating Virtual MFA Device authentication code 2: %s", err)
		}

		enableVirtualMFADeviceInput := &iam.EnableMFADeviceInput{
			AuthenticationCode1: aws.String(authenticationCode1),
			AuthenticationCode2: aws.String(authenticationCode2),
			SerialNumber:        createVirtualMFADeviceOutput.VirtualMFADevice.SerialNumber,
			UserName:            getUserOutput.User.UserName,
		}

		if _, err := conn.EnableMFADevice(enableVirtualMFADeviceInput); err != nil {
			return fmt.Errorf("error enabling IAM User (%s) Virtual MFA Device: %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

// Creates an IAM User SSH Key outside of Terraform to verify that it is deleted when `force_destroy` is set
func testAccCheckUserUploadsSSHKey(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		publicKey, _, err := RandSSHKeyPairSize(2048, acctest.DefaultEmailAddress)
		if err != nil {
			return fmt.Errorf("error generating random SSH key: %w", err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		input := &iam.UploadSSHPublicKeyInput{
			UserName:         getUserOutput.User.UserName,
			SSHPublicKeyBody: aws.String(publicKey),
		}

		_, err = conn.UploadSSHPublicKey(input)
		if err != nil {
			return fmt.Errorf("error uploading IAM User (%s) SSH key: %w", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

func testAccCheckUserUploadSigningCertificate(getUserOutput *iam.GetUserOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		signingCertificate, err := os.ReadFile("./test-fixtures/iam-ssl-unix-line-endings.pem")
		if err != nil {
			return fmt.Errorf("error reading signing certificate fixture: %s", err)
		}
		input := &iam.UploadSigningCertificateInput{
			CertificateBody: aws.String(string(signingCertificate)),
			UserName:        getUserOutput.User.UserName,
		}

		if _, err := conn.UploadSigningCertificate(input); err != nil {
			return fmt.Errorf("error uploading IAM User (%s) Signing Certificate : %s", aws.StringValue(getUserOutput.User.UserName), err)
		}

		return nil
	}
}

func testAccUserConfig(rName, path string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name = %q
  path = %q
}
`, rName, path)
}

func testAccUserConfig_permissionsBoundary(rName, permissionsBoundary string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user" {
  name                 = %q
  permissions_boundary = %q
}
`, rName, permissionsBoundary)
}

func testAccUserForceDestroyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  force_destroy = true
  name          = %q
}
`, rName)
}

func testAccUserConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %q

  tags = {
    Name = "test-Name"
    tag2 = "test-tag2"
  }
}
`, rName)
}

func testAccUserConfig_tagsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %q

  tags = {
    tag2 = "test-tagUpdate"
  }
}
`, rName)
}
