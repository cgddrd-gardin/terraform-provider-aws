package efs_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfefs "github.com/hashicorp/terraform-provider-aws/internal/service/efs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_efs_mount_target", &resource.Sweeper{
		Name: "aws_efs_mount_target",
		F:    sweepMountTargets,
	})
}

func sweepMountTargets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EFSConn

	var errors error
	input := &efs.DescribeFileSystemsInput{}
	err = conn.DescribeFileSystemsPages(input, func(page *efs.DescribeFileSystemsOutput, lastPage bool) bool {
		for _, filesystem := range page.FileSystems {
			id := aws.StringValue(filesystem.FileSystemId)
			log.Printf("[INFO] Deleting Mount Targets for EFS File System: %s", id)

			var errors error
			input := &efs.DescribeMountTargetsInput{
				FileSystemId: filesystem.FileSystemId,
			}
			for {
				out, err := conn.DescribeMountTargets(input)
				if err != nil {
					errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS Mount Targets on File System %q: %w", id, err))
					break
				}

				if out == nil || len(out.MountTargets) == 0 {
					log.Printf("[INFO] No EFS Mount Targets to sweep on File System %q", id)
					break
				}

				for _, mounttarget := range out.MountTargets {
					id := aws.StringValue(mounttarget.MountTargetId)

					log.Printf("[INFO] Deleting EFS Mount Target: %s", id)
					_, err := conn.DeleteMountTarget(&efs.DeleteMountTargetInput{
						MountTargetId: mounttarget.MountTargetId,
					})
					if err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error deleting EFS Mount Target %q: %w", id, err))
						continue
					}

					err = tfefs.WaitForDeleteMountTarget(conn, id, 10*time.Minute)
					if err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error waiting for EFS Mount Target %q to delete: %w", id, err))
						continue
					}
				}

				if out.NextMarker == nil {
					break
				}
				input.Marker = out.NextMarker
			}
		}
		return true
	})
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS File Systems: %w", err))
	}

	return errors
}

func TestAccEFSMountTarget_basic(t *testing.T) {
	var mount efs.MountTargetDescription
	ct := fmt.Sprintf("createtoken-%d", sdkacctest.RandInt())
	resourceName := "aws_efs_mount_target.test"
	resourceName2 := "aws_efs_mount_target.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, efs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEfsMountTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig(ct),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsMountTarget(resourceName, &mount),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrSet(resourceName, "availability_zone_name"),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					acctest.MatchResourceAttrRegionalARN(resourceName, "file_system_arn", "elasticfilesystem", regexp.MustCompile(`file-system/fs-.+`)),
					resource.TestMatchResourceAttr(resourceName, "ip_address", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
					resource.TestCheckResourceAttrSet(resourceName, "mount_target_dns_name"),
					resource.TestCheckResourceAttrSet(resourceName, "network_interface_id"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMountTargetModifiedConfig(ct),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsMountTarget(resourceName, &mount),
					testAccCheckEfsMountTarget(resourceName2, &mount),
					acctest.MatchResourceAttrRegionalHostname(resourceName, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
					acctest.MatchResourceAttrRegionalHostname(resourceName2, "dns_name", "efs", regexp.MustCompile(`fs-[^.]+`)),
				),
			},
		},
	})
}

func TestAccEFSMountTarget_disappears(t *testing.T) {
	var mount efs.MountTargetDescription
	resourceName := "aws_efs_mount_target.test"
	ct := fmt.Sprintf("createtoken-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, efs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpnGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetConfig(ct),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsMountTarget(resourceName, &mount),
					acctest.CheckResourceDisappears(acctest.Provider, tfefs.ResourceMountTarget(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEFSMountTarget_ipAddress(t *testing.T) {
	var mount efs.MountTargetDescription
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, efs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEfsMountTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetIPAddressConfig("10.0.0.100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsMountTarget(resourceName, &mount),
					resource.TestCheckResourceAttr(resourceName, "ip_address", "10.0.0.100"),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13845
func TestAccEFSMountTarget_IPAddress_emptyString(t *testing.T) {
	var mount efs.MountTargetDescription
	resourceName := "aws_efs_mount_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, efs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEfsMountTargetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMountTargetIPAddressConfig(""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsMountTarget(resourceName, &mount),
					resource.TestMatchResourceAttr(resourceName, "ip_address", regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`)),
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

func TestMountTarget_hasEmptyMountTargets(t *testing.T) {
	mto := &efs.DescribeMountTargetsOutput{
		MountTargets: []*efs.MountTargetDescription{},
	}

	actual := tfefs.HasEmptyMountTargets(mto)
	if !actual {
		t.Fatalf("Expected return value to be true, got %t", actual)
	}

	// Add an empty mount target.
	mto.MountTargets = append(mto.MountTargets, &efs.MountTargetDescription{})

	actual = tfefs.HasEmptyMountTargets(mto)
	if actual {
		t.Fatalf("Expected return value to be false, got %t", actual)
	}
}

func testAccCheckEfsMountTargetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_mount_target" {
			continue
		}

		resp, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
			MountTargetId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			if tfawserr.ErrMessageContains(err, efs.ErrCodeMountTargetNotFound, "") {
				// gone
				return nil
			}
			return fmt.Errorf("Error describing EFS Mount in tests: %s", err)
		}
		if len(resp.MountTargets) > 0 {
			return fmt.Errorf("EFS Mount target %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckEfsMountTarget(resourceID string, mount *efs.MountTargetDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		fs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceID)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn
		mt, err := conn.DescribeMountTargets(&efs.DescribeMountTargetsInput{
			MountTargetId: aws.String(fs.Primary.ID),
		})
		if err != nil {
			return err
		}

		if aws.StringValue(mt.MountTargets[0].MountTargetId) != fs.Primary.ID {
			return fmt.Errorf("Mount target ID mismatch: %q != %q",
				*mt.MountTargets[0].MountTargetId, fs.Primary.ID)
		}

		*mount = *mt.MountTargets[0]

		return nil
	}
}

func testAccMountTargetConfig(ct string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_efs_file_system" "test" {
  creation_token = "%s"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}
`, ct)
}

func testAccMountTargetModifiedConfig(ct string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_efs_file_system" "test" {
  creation_token = "%s"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test.id
}

resource "aws_efs_mount_target" "test2" {
  file_system_id = aws_efs_file_system.test.id
  subnet_id      = aws_subnet.test2.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test2"
  }
}
`, ct)
}

func testAccMountTargetIPAddressConfig(ipAddress string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"

  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_efs_file_system" "test" {
  tags = {
    Name = "tf-acc-efs-mount-target-test"
  }
}

resource "aws_efs_mount_target" "test" {
  file_system_id = aws_efs_file_system.test.id
  ip_address     = %[1]q
  subnet_id      = aws_subnet.test.id
}
`, ipAddress)
}

func testAccCheckVpnGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpn_gateway" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{
			VpnGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			var v *ec2.VpnGateway
			for _, g := range resp.VpnGateways {
				if *g.VpnGatewayId == rs.Primary.ID {
					v = g
				}
			}

			if v == nil {
				// wasn't found
				return nil
			}

			if *v.State != "deleted" {
				return fmt.Errorf("Expected VPN Gateway to be in deleted state, but was not: %s", v)
			}
			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidVpnGatewayID.NotFound" {
			return err
		}
	}

	return nil
}
