package cloudfront_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloudfront_key_group", &resource.Sweeper{
		Name: "aws_cloudfront_key_group",
		F:    sweepKeyGroup,
	})
}

func sweepKeyGroup(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).CloudFrontConn
	var sweeperErrs *multierror.Error

	input := &cloudfront.ListKeyGroupsInput{}

	for {
		output, err := conn.ListKeyGroups(input)
		if err != nil {
			if sweep.SkipSweepError(err) {
				log.Printf("[WARN] Skipping CloudFront key group sweep for %s: %s", region, err)
				return nil
			}
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudFront key group: %w", err))
			return sweeperErrs.ErrorOrNil()
		}

		if output == nil || output.KeyGroupList == nil || len(output.KeyGroupList.Items) == 0 {
			log.Print("[DEBUG] No CloudFront key group to sweep")
			return nil
		}

		for _, item := range output.KeyGroupList.Items {
			strId := aws.StringValue(item.KeyGroup.Id)
			log.Printf("[INFO] CloudFront key group %s", strId)
			_, err := conn.DeleteKeyGroup(&cloudfront.DeleteKeyGroupInput{
				Id: item.KeyGroup.Id,
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CloudFront key group %s: %w", strId, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if output.KeyGroupList.NextMarker == nil {
			break
		}
		input.Marker = output.KeyGroupList.NextMarker
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccCloudFrontKeyGroup_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_key_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontKeyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "comment", "test key group"),
					resource.TestCheckResourceAttrSet("aws_cloudfront_key_group.test", "etag"),
					resource.TestCheckResourceAttrSet("aws_cloudfront_key_group.test", "id"),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "name", rName),
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

func TestAccCloudFrontKeyGroup_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_key_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontKeyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcloudfront.ResourceKeyGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccCloudFrontKeyGroup_comment(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_key_group.test"

	firstComment := "first comment"
	secondComment := "second comment"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontKeyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupCommentConfig(rName, firstComment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "comment", firstComment),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyGroupCommentConfig(rName, secondComment),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "comment", secondComment),
				),
			},
		},
	})
}

func TestAccCloudFrontKeyGroup_items(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudfront_key_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(cloudfront.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, cloudfront.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCloudFrontKeyGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKeyGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKeyGroupItemsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontKeyGroupExistence(resourceName),
					resource.TestCheckResourceAttr("aws_cloudfront_key_group.test", "items.#", "2"),
				),
			},
		},
	})
}

func testAccCheckCloudFrontKeyGroupExistence(r string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[r]
		if !ok {
			return fmt.Errorf("not found: %s", r)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no Id is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

		input := &cloudfront.GetKeyGroupInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetKeyGroup(input)
		if err != nil {
			return fmt.Errorf("error retrieving CloudFront key group: %s", err)
		}
		return nil
	}
}

func testAccCheckCloudFrontKeyGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).CloudFrontConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_key_group" {
			continue
		}

		input := &cloudfront.GetKeyGroupInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetKeyGroup(input)
		if tfawserr.ErrMessageContains(err, cloudfront.ErrCodeNoSuchResource, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("CloudFront key group (%s) was not deleted", rs.Primary.ID)
	}

	return nil
}

func testAccKeyGroupBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test" {
  comment     = "test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = %q
}
`, rName)
}

func testAccKeyGroupConfig(rName string) string {
	return testAccKeyGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cloudfront_key_group" "test" {
  comment = "test key group"
  items   = [aws_cloudfront_public_key.test.id]
  name    = %q
}
`, rName)
}

func testAccKeyGroupCommentConfig(rName string, comment string) string {
	return testAccKeyGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cloudfront_key_group" "test" {
  comment = %q
  items   = [aws_cloudfront_public_key.test.id]
  name    = %q
}
`, comment, rName)
}

func testAccKeyGroupItemsConfig(rName string) string {
	return testAccKeyGroupBaseConfig(rName) + fmt.Sprintf(`
resource "aws_cloudfront_public_key" "test2" {
  comment     = "second test key"
  encoded_key = file("test-fixtures/cloudfront-public-key.pem")
  name        = "%[1]s-second"
}

resource "aws_cloudfront_key_group" "test" {
  comment = "test key group"
  items   = [aws_cloudfront_public_key.test.id, aws_cloudfront_public_key.test2.id]
  name    = %[1]q
}
`, rName)
}
