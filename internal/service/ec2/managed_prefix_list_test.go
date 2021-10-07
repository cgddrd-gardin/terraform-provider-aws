package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccAwsEc2ManagedPrefixList_basic(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Name(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "address_family", "IPv4"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`prefix-list/pl-[[:xdigit:]]+`)),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "max_entries", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListConfigUpdated(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "max_entries", "2"),
				),
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixList_disappears(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListConfig_Name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceManagedPrefixList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixList_AddressFamily_IPv6(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_AddressFamily(rName, "IPv6"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "address_family", "IPv6"),
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

func TestAccAwsEc2ManagedPrefixList_Entry_Cidr(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Entry_Cidr1(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "1.0.0.0/8",
						"description": "Test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "2.0.0.0/8",
						"description": "Test2",
					}),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Entry_Cidr2(rName),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "1.0.0.0/8",
						"description": "Test1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "3.0.0.0/8",
						"description": "Test3",
					}),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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

func TestAccAwsEc2ManagedPrefixList_Entry_Description(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Entry_Description(rName, "description1"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "1.0.0.0/8",
						"description": "description1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "2.0.0.0/8",
						"description": "description1",
					}),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Entry_Description(rName, "description2"),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "entry.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "1.0.0.0/8",
						"description": "description2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "entry.*", map[string]string{
						"cidr":        "2.0.0.0/8",
						"description": "description2",
					}),
					resource.TestCheckResourceAttr(resourceName, "version", "3"), // description-only updates require two operations
				),
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixList_Name(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName1 := sdkacctest.RandomWithPrefix("tf-acc-test")
	rName2 := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Name(rName1),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName1),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:       testAccAwsEc2ManagedPrefixListConfig_Name(rName2),
				ResourceName: resourceName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName2),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

func TestAccAwsEc2ManagedPrefixList_Tags(t *testing.T) {
	resourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAwsEc2ManagedPrefixListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2ManagedPrefixListConfig_Tags1(rName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsEc2ManagedPrefixListConfig_Tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccAwsEc2ManagedPrefixListConfig_Tags1(rName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccAwsEc2ManagedPrefixListExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

func testAccCheckAwsEc2ManagedPrefixListDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_managed_prefix_list" {
			continue
		}

		_, err := tfec2.FindManagedPrefixListByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EC2 Managed Prefix List %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccAwsEc2ManagedPrefixListExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Managed Prefix List ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		_, err := tfec2.FindManagedPrefixListByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccPreCheckEc2ManagedPrefixList(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeManagedPrefixListsInput{}

	_, err := conn.DescribeManagedPrefixLists(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAwsEc2ManagedPrefixListConfig_AddressFamily(rName string, addressFamily string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = %[2]q
  max_entries    = 1
  name           = %[1]q
}
`, rName, addressFamily)
}

func testAccAwsEc2ManagedPrefixListConfig_Entry_Cidr1(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 5
  name           = %[1]q

  entry {
    cidr        = "1.0.0.0/8"
    description = "Test1"
  }

  entry {
    cidr        = "2.0.0.0/8"
    description = "Test2"
  }
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_Entry_Cidr2(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 5
  name           = %[1]q

  entry {
    cidr        = "1.0.0.0/8"
    description = "Test1"
  }

  entry {
    cidr        = "3.0.0.0/8"
    description = "Test3"
  }
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_Entry_Description(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 5
  name           = %[1]q

  entry {
    cidr        = "1.0.0.0/8"
    description = %[2]q
  }

  entry {
    cidr        = "2.0.0.0/8"
    description = %[2]q
  }
}
`, rName, description)
}

func testAccAwsEc2ManagedPrefixListConfig_Name(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 2
  name           = %[1]q
}
`, rName)
}

func testAccAwsEc2ManagedPrefixListConfig_Tags1(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAwsEc2ManagedPrefixListConfig_Tags2(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 5

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}