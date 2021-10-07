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
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfdatasync "github.com/hashicorp/terraform-provider-aws/internal/service/datasync"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_datasync_location_smb", &resource.Sweeper{
		Name: "aws_datasync_location_smb",
		F:    sweepLocationSMBs,
	})
}

func sweepLocationSMBs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).DataSyncConn

	input := &datasync.ListLocationsInput{}
	for {
		output, err := conn.ListLocations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DataSync Location SMB sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving DataSync Location SMBs: %w", err)
		}

		if len(output.Locations) == 0 {
			log.Print("[DEBUG] No DataSync Location SMBs to sweep")
			return nil
		}

		for _, location := range output.Locations {
			uri := aws.StringValue(location.LocationUri)
			if !strings.HasPrefix(uri, "smb://") {
				log.Printf("[INFO] Skipping DataSync Location SMB: %s", uri)
				continue
			}
			log.Printf("[INFO] Deleting DataSync Location SMB: %s", uri)

			r := tfdatasync.ResourceLocationSMB()
			d := r.Data(nil)
			d.SetId(aws.StringValue(location.LocationArn))
			err = r.Delete(d, client)
			if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
				continue
			}

			if err != nil {
				log.Printf("[ERROR] Failed to delete DataSync Location SMB (%s): %s", uri, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSDataSyncLocationSmb_basic(t *testing.T) {
	var locationSmb1 datasync.DescribeLocationSmbOutput

	resourceName := "aws_datasync_location_smb.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationSMBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationSMBConfig(rName, "/test/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(resourceName, &locationSmb1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "user", "Guest"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^smb://.+/`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "server_hostname"},
			},
			{
				Config: testAccLocationSMBConfig(rName, "/test2/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(resourceName, &locationSmb1),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "datasync", regexp.MustCompile(`location/loc-.+`)),
					resource.TestCheckResourceAttr(resourceName, "agent_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mount_options.0.version", "AUTOMATIC"),
					resource.TestCheckResourceAttr(resourceName, "user", "Guest"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestMatchResourceAttr(resourceName, "uri", regexp.MustCompile(`^smb://.+/`)),
				),
			},
		},
	})
}

func TestAccAWSDataSyncLocationSmb_disappears(t *testing.T) {
	var locationSmb1 datasync.DescribeLocationSmbOutput
	resourceName := "aws_datasync_location_smb.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationSMBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationSMBConfig(rName, "/test/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(resourceName, &locationSmb1),
					testAccCheckLocationSMBDisappears(&locationSmb1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDataSyncLocationSmb_Tags(t *testing.T) {
	var locationSmb1, locationSmb2, locationSmb3 datasync.DescribeLocationSmbOutput
	resourceName := "aws_datasync_location_smb.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, datasync.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLocationSMBDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLocationSMBTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(resourceName, &locationSmb1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "server_hostname"},
			},
			{
				Config: testAccLocationSMBTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(resourceName, &locationSmb2),
					testAccCheckLocationSMBNotRecreated(&locationSmb1, &locationSmb2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccLocationSMBTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLocationSMBExists(resourceName, &locationSmb3),
					testAccCheckLocationSMBNotRecreated(&locationSmb2, &locationSmb3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccCheckLocationSMBDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_datasync_location_smb" {
			continue
		}

		input := &datasync.DescribeLocationSmbInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeLocationSmb(input)

		if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
			return nil
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func testAccCheckLocationSMBExists(resourceName string, locationSmb *datasync.DescribeLocationSmbOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn
		input := &datasync.DescribeLocationSmbInput{
			LocationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeLocationSmb(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Location %q does not exist", rs.Primary.ID)
		}

		*locationSmb = *output

		return nil
	}
}

func testAccCheckLocationSMBDisappears(location *datasync.DescribeLocationSmbOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DataSyncConn

		input := &datasync.DeleteLocationInput{
			LocationArn: location.LocationArn,
		}

		_, err := conn.DeleteLocation(input)

		return err
	}
}

func testAccCheckLocationSMBNotRecreated(i, j *datasync.DescribeLocationSmbOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.TimeValue(i.CreationTime).Equal(aws.TimeValue(j.CreationTime)) {
			return errors.New("DataSync Location SMB was recreated")
		}

		return nil
	}
}

func testAccLocationSMBBaseConfig(rName string) string {
	return acctest.ConfigCompose(
		// Reference: https://docs.aws.amazon.com/datasync/latest/userguide/agent-requirements.html
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "m5.2xlarge", "m5.4xlarge"),
		fmt.Sprintf(`
data "aws_partition" "current" {}

# Reference: https://docs.aws.amazon.com/datasync/latest/userguide/deploy-agents.html
data "aws_ssm_parameter" "aws_service_datasync_ami" {
  name = "/aws/service/datasync/ami"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-datasync-location-smb"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-datasync-location-smb"
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-datasync-location-smb"
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = "tf-acc-test-datasync-location-smb"
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "tf-acc-test-datasync-smb"
  }
}

resource "aws_instance" "test" {
  depends_on = [aws_default_route_table.test]

  ami                         = data.aws_ssm_parameter.aws_service_datasync_ami.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = "tf-acc-test-datasync-smb"
  }
}

resource "aws_datasync_agent" "test" {
  ip_address = aws_instance.test.public_ip
  name       = %[1]q
}
`, rName))
}

func testAccLocationSMBConfig(rName, dir string) string {
	return testAccLocationSMBBaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  password        = "ZaphodBeeblebroxPW"
  server_hostname = aws_instance.test.public_ip
  subdirectory    = %[1]q
  user            = "Guest"
}
`, dir)
}

func testAccLocationSMBTags1Config(rName, key1, value1 string) string {
	return testAccLocationSMBBaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  password        = "ZaphodBeeblebroxPW"
  server_hostname = aws_instance.test.public_ip
  subdirectory    = "/test/"
  user            = "Guest"

  tags = {
    %[1]q = %[2]q
  }
}
`, key1, value1)
}

func testAccLocationSMBTags2Config(rName, key1, value1, key2, value2 string) string {
	return testAccLocationSMBBaseConfig(rName) + fmt.Sprintf(`
resource "aws_datasync_location_smb" "test" {
  agent_arns      = [aws_datasync_agent.test.arn]
  password        = "ZaphodBeeblebroxPW"
  server_hostname = aws_instance.test.public_ip
  subdirectory    = "/test/"
  user            = "Guest"

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, key1, value1, key2, value2)
}
