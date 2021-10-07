package datasync_test

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_datasync_location_efs", &resource.Sweeper{
		Name: "aws_datasync_location_efs",
		F:    sweepLocationEFSs,
	})
}

func sweepLocationEFSs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location EFS sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location EFSs: %s", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location EFSs to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "efs://") {
				log.Printf("[INFO] Skipping DataSync Location EFS: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location EFS: %s", uri)
			input := &datasync.DeleteLocationInput{
				LocationArn: location.LocationArn,
			}

			_, err := conn.DeleteLocation(input)

			if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location EFS (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSDataSyncLocationEfs_basic(t *testing.T) {
	var locationEfs1 datasync.DescribeLocationEfsOutput
	efsFileSystemResourceName := "aws_efs_file_system.test"
	resourceName := "aws_datasync_location_efs.test"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationEFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "ec2_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ec2_config.0.security_group_arns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_config.0.subnet_arn", subnetResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "efs_file_system_arn", efsFileSystemResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^efs://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"efs_file_system_arn"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationEfs_disappears(t *testing.T) {
	var locationEfs1 datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationEFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs1),
					testAccCheckLocationEFSDisappears(&locationEfs1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncLocationEfs_Subdirectory(t *testing.T) {
	var locationEfs1 datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationEFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSSubdirectoryConfig("/subdirectory1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs1),
					resource.TestCheckResourceAttr(resourceName, "subdirectory", "/subdirectory1/"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"efs_file_system_arn"},
			},
		},
	})
}

func TestAccAWSDataSyncLocationEfs_Tags(t *testing.T) {
	var locationEfs1, locationEfs2, locationEfs3 datasync.DescribeLocationEfsOutput
	resourceName := "aws_datasync_location_efs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationEFSDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationEFSTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"efs_file_system_arn"},
			},
			{
				Config: testAccLocationEFSTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs2),
					testAccCheckLocationEFSNotRecreated(&locationEfs1, &locationEfs2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationEFSTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationEFSExists(resourceName, &locationEfs3),
					testAccCheckLocationEFSNotRecreated(&locationEfs2, &locationEfs3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationEFSDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_efs" {
			continue
		}

		input := &datasync.DescribeLocationEfsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationEfs(input)

		if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckLocationEFSExists(resourceName string, locationEfs *datasync.DescribeLocationEfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn
		input := &datasync.DescribeLocationEfsInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationEfs(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationEfs = *output

		return nil
	}
}

func testAccCheckLocationEFSDisappears(location *datasync.DescribeLocationEfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

		input := &datasync.DeleteLocationInput{
			LocationArn: location.LocationArn,
		}

		_, err := conn.DeleteLocation(input)

		return err
	}
}

func testAccCheckLocationEFSNotRecreated(i, j *datasync.DescribeLocationEfsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("DataSync Location EFS was recreated")
		}

		return nil
	}
}

func testAccLocationEFSBaseConfig() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-datasync-location-efs"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-datasync-location-efs"
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-datasync-location-efs"
  }
}

resource "aws_efs_file_system" "test" {}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}
`
}

func testAccLocationEFSConfig() string {
	return testAccLocationEFSBaseConfig() + `
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }
}
`
}

func testAccLocationEFSSubdirectoryConfig(subdirectory string) string {
	return testAccLocationEFSBaseConfig() + fmt.Sprintf(`
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn
  subdirectory        = %q

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }
}
`, subdirectory)
}

func testAccLocationEFSTags1Config(key1, value1 string) string {
	return testAccLocationEFSBaseConfig() + fmt.Sprintf(`
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }

  tags = {
    %q = %q
  }
}
`, key1, value1)
}

func testAccLocationEFSTags2Config(key1, value1, key2, value2 string) string {
	return testAccLocationEFSBaseConfig() + fmt.Sprintf(`
resource "aws_datasync_location_efs" "test" {
  efs_file_system_arn = aws_efs_mount_target.test.file_system_arn

  ec2_config {
    security_group_arns = [aws_security_group.test.arn]
    subnet_arn          = aws_subnet.test.arn
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, key1, value1, key2, value2)
}
