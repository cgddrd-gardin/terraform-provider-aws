package iot_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iot"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_iot_certificate", &resource.Sweeper{
		Name: "aws_iot_certificate",
		F:    sweepCertifcates,
		Dependencies: []string{
			"aws_iot_policy_attachment",
			"aws_iot_thing_principal_attachment",
		},
	})
}

func sweepCertifcates(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).IoTConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &iot.ListCertificatesInput{}

	err = conn.ListCertificatesPages(input, func(page *iot.ListCertificatesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, certificate := range page.Certificates {
			r := tfiot.ResourceCertificate()
			d := r.Data(nil)

			d.SetId(aws.StringValue(certificate.CertificateId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing IoT Certificate for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping IoT Certificate for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping IoT Certificate sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSIoTCertificate_csr(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificate_csr,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "arn"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "csr"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "certificate_pem"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "public_key"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "private_key"),
					resource.TestCheckResourceAttr("aws_iot_certificate.foo_cert", "active", "true"),
				),
			},
		},
	})
}

func TestAccAWSIoTCertificate_keys_certificate(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iot.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificate_keys_certificate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "arn"),
					resource.TestCheckNoResourceAttr("aws_iot_certificate.foo_cert", "csr"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "certificate_pem"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "public_key"),
					resource.TestCheckResourceAttrSet("aws_iot_certificate.foo_cert", "private_key"),
					resource.TestCheckResourceAttr("aws_iot_certificate.foo_cert", "active", "true"),
				),
			},
		},
	})
}

func testAccCheckCertificateDestroy_basic(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IoTConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_certificate" {
			continue
		}

		// Try to find the Cert
		DescribeCertOpts := &iot.DescribeCertificateInput{
			CertificateId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeCertificate(DescribeCertOpts)

		if err == nil {
			if resp.CertificateDescription != nil {
				return fmt.Errorf("Device Certificate still exists")
			}
		}

		// Verify the error is what we want
		if err != nil {
			iotErr, ok := err.(awserr.Error)
			if !ok || iotErr.Code() != "ResourceNotFoundException" {
				return err
			}
		}

	}

	return nil
}

var testAccCertificate_csr = `
resource "aws_iot_certificate" "foo_cert" {
  csr    = file("test-fixtures/iot-csr.pem")
  active = true
}
`

var testAccCertificate_keys_certificate = `
resource "aws_iot_certificate" "foo_cert" {
  active = true
}
`
