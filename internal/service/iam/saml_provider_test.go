package iam_test

import (
	"fmt"
	"log"
	"testing"

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
)

func init() {
	resource.AddTestSweepers("aws_iam_saml_provider", &resource.Sweeper{
		Name: "aws_iam_saml_provider",
		F:    sweepSamlProvider,
	})
}

func sweepSamlProvider(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).IAMConn

	var sweeperErrs *multierror.Error

	out, err := conn.ListSAMLProviders(&iam.ListSAMLProvidersInput{})

	for _, sampProvider := range out.SAMLProviderList {
		arn := aws.StringValue(sampProvider.Arn)

		r := tfiam.ResourceSamlProvider()
		d := r.Data(nil)
		d.SetId(arn)
		err := r.Delete(d, client)

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting IAM SAML Provider (%s): %w", arn, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping IAM SAML Provider sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error describing IAM SAML Providers: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccIAMSamlProvider_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMSamlProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMSamlProviderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("saml-provider/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_metadata_document"),
					resource.TestCheckResourceAttrSet(resourceName, "valid_until"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccIAMSamlProviderConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "saml_metadata_document"),
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

func TestAccIAMSamlProvider_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMSamlProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMSamlProviderConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMSamlProviderConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccIAMSamlProviderConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIAMSamlProvider_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIAMSamlProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMSamlProviderConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMSamlProviderExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourceSamlProvider(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIAMSamlProviderDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_saml_provider" {
			continue
		}

		input := &iam.GetSAMLProviderInput{
			SAMLProviderArn: aws.String(rs.Primary.ID),
		}
		out, err := conn.GetSAMLProvider(input)

		if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if out != nil {
			return fmt.Errorf("IAM SAML Provider (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckIAMSamlProviderExists(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("Not Found: %s", id)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn
		_, err := conn.GetSAMLProvider(&iam.GetSAMLProviderInput{
			SAMLProviderArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccIAMSamlProviderConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %q
  saml_metadata_document = file("./test-fixtures/saml-metadata.xml")
}
`, rName)
}

func testAccIAMSamlProviderConfigUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %q
  saml_metadata_document = file("./test-fixtures/saml-metadata-modified.xml")
}
`, rName)
}

func testAccIAMSamlProviderConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %q
  saml_metadata_document = file("./test-fixtures/saml-metadata.xml")

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccIAMSamlProviderConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %q
  saml_metadata_document = file("./test-fixtures/saml-metadata.xml")

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
