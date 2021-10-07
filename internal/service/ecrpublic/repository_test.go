package ecrpublic_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecrpublic "github.com/hashicorp/terraform-provider-aws/internal/service/ecrpublic"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_ecrpublic_repository", &resource.Sweeper{
		Name: "aws_ecrpublic_repository",
		F:    sweepRepositories,
	})
}

func sweepRepositories(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).ECRPublicConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	err = conn.DescribeRepositoriesPages(&ecrpublic.DescribeRepositoriesInput{}, func(page *ecrpublic.DescribeRepositoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, repository := range page.Repositories {
			r := tfecrpublic.ResourceRepository()
			d := r.Data(nil)
			d.SetId(aws.StringValue(repository.RepositoryName))
			d.Set("registry_id", repository.RegistryId)
			d.Set("force_destroy", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing ECR Public Repositories for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping ECR Public Repositories for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping ECR Public Repositories sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSEcrPublicRepository_basic(t *testing.T) {
	var v ecrpublic.Repository
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "repository_name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "ecr-public", "repository/"+rName),
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

func TestAccAWSEcrPublicRepository_catalogdata_abouttext(t *testing.T) {
	var v ecrpublic.Repository
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCatalogDataAboutTextConfig(rName, "about_text_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.about_text", "about_text_1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryCatalogDataAboutTextConfig(rName, "about_text_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.about_text", "about_text_2"),
				),
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_catalogdata_architectures(t *testing.T) {
	var v ecrpublic.Repository
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCatalogDataArchitecturesConfig(rName, "Linux"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.0", "Linux"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryCatalogDataArchitecturesConfig(rName, "Windows"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.architectures.0", "Windows"),
				),
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_catalogdata_description(t *testing.T) {
	var v ecrpublic.Repository
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCatalogDataDescriptionConfig(rName, "description 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.description", "description 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryCatalogDataDescriptionConfig(rName, "description 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.description", "description 2"),
				),
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_catalogdata_operatingsystems(t *testing.T) {
	var v ecrpublic.Repository
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCatalogDataOperatingSystemsConfig(rName, "ARM"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.0", "ARM"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryCatalogDataOperatingSystemsConfig(rName, "x86"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.operating_systems.0", "x86"),
				),
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_catalogdata_usagetext(t *testing.T) {
	var v ecrpublic.Repository
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCatalogDataUsageTextConfig(rName, "usage text 1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.usage_text", "usage text 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryCatalogDataUsageTextConfig(rName, "usage text 2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.0.usage_text", "usage text 2"),
				),
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_catalogdata_logoimageblob(t *testing.T) {
	var v ecrpublic.Repository
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryCatalogDataLogoImageBlobConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "catalog_data.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "catalog_data.0.logo_image_blob"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"catalog_data.0.logo_image_blob"},
			},
		},
	})
}

func TestAccAWSEcrPublicRepository_basic_forcedestroy(t *testing.T) {
	var v ecrpublic.Repository
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryForceDestroyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "repository_name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "registry_id"),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "ecr-public", "repository/"+rName),
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

func TestAccAWSEcrPublicRepository_disappears(t *testing.T) {
	var v ecrpublic.Repository
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ecrpublic_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ecrpublic.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfecrpublic.ResourceRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepositoryExists(name string, res *ecrpublic.Repository) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ECR Public repository ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECRPublicConn

		output, err := conn.DescribeRepositories(&ecrpublic.DescribeRepositoriesInput{
			RepositoryNames: aws.StringSlice([]string{rs.Primary.ID}),
		})
		if err != nil {
			return err
		}
		if len(output.Repositories) == 0 {
			return fmt.Errorf("ECR Public repository %s not found", rs.Primary.ID)
		}

		*res = *output.Repositories[0]

		return nil
	}
}

func testAccCheckRepositoryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECRPublicConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecrpublic_repository" {
			continue
		}

		input := ecrpublic.DescribeRepositoriesInput{
			RepositoryNames: []*string{aws.String(rs.Primary.Attributes["repository_name"])},
		}

		out, err := conn.DescribeRepositories(&input)

		if tfawserr.ErrMessageContains(err, ecrpublic.ErrCodeRepositoryNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		for _, repository := range out.Repositories {
			if aws.StringValue(repository.RepositoryName) == rs.Primary.Attributes["repository_name"] {
				return fmt.Errorf("ECR Public repository still exists: %s", rs.Primary.Attributes["repository_name"])
			}
		}
	}

	return nil
}

func testAccRepositoryConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
}
`, rName)
}

func testAccRepositoryForceDestroyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  force_destroy   = true
}
`, rName)
}

func testAccRepositoryCatalogDataAboutTextConfig(rName string, aboutText string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    about_text = %[2]q
  }
}
`, rName, aboutText)
}

func testAccRepositoryCatalogDataArchitecturesConfig(rName string, architecture string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    architectures = [%[2]q]
  }
}
`, rName, architecture)
}

func testAccRepositoryCatalogDataDescriptionConfig(rName string, description string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    description = %[2]q
  }
}
`, rName, description)
}

func testAccRepositoryCatalogDataOperatingSystemsConfig(rName string, operatingSystem string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    operating_systems = [%[2]q]
  }
}
`, rName, operatingSystem)
}

func testAccRepositoryCatalogDataUsageTextConfig(rName string, usageText string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %[1]q
  catalog_data {
    usage_text = %[2]q
  }
}
`, rName, usageText)
}

func testAccRepositoryCatalogDataLogoImageBlobConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecrpublic_repository" "test" {
  repository_name = %q
  catalog_data {
    logo_image_blob = filebase64("test-fixtures/terraform_logo.png")
  }
}
`, rName)
}

func testAccPreCheck(t *testing.T) {
	// At this time, calls to DescribeRepositories returns (and by default, retries)
	// an InternalFailure when the region is not supported i.e. not us-east-1.
	// TODO: Remove when ECRPublic is supported across other known regions
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/18047
	if region := acctest.Provider.Meta().(*conns.AWSClient).Region; region != endpoints.UsEast1RegionID {
		t.Skipf("skipping acceptance testing: region (%s) does not support ECR Public repositories", region)
	}

	conn := acctest.Provider.Meta().(*conns.AWSClient).ECRPublicConn
	input := &ecrpublic.DescribeRepositoriesInput{}
	_, err := conn.DescribeRepositories(input)
	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
