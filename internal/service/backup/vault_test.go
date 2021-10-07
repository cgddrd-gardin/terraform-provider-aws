package backup_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfbackup "github.com/hashicorp/terraform-provider-aws/internal/service/backup"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_backup_vault", &resource.Sweeper{
		Name: "aws_backup_vault",
		F:    sweepVaults,
		Dependencies: []string{
			"aws_backup_vault_notifications",
			"aws_backup_vault_policy",
		},
	})
}

func sweepVaults(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).BackupConn
	input := &backup.ListBackupVaultsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.ListBackupVaultsPages(input, func(page *backup.ListBackupVaultsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, vault := range page.BackupVaultList {
			failedToDeleteRecoveryPoint := false
			name := aws.StringValue(vault.BackupVaultName)
			input := &backup.ListRecoveryPointsByBackupVaultInput{
				BackupVaultName: aws.String(name),
			}

			err := conn.ListRecoveryPointsByBackupVaultPages(input, func(page *backup.ListRecoveryPointsByBackupVaultOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, recoveryPoint := range page.RecoveryPoints {
					arn := aws.StringValue(recoveryPoint.RecoveryPointArn)

					log.Printf("[INFO] Deleting Recovery Point (%s) in Backup Vault (%s)", arn, name)
					_, err := conn.DeleteRecoveryPoint(&backup.DeleteRecoveryPointInput{
						BackupVaultName:  aws.String(name),
						RecoveryPointArn: aws.String(arn),
					})

					if err != nil {
						log.Printf("[WARN] Failed to delete Recovery Point (%s) in Backup Vault (%s): %s", arn, name, err)
						failedToDeleteRecoveryPoint = true
					}
				}

				return !lastPage
			})

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Reovery Points in Backup Vault (%s) for %s: %w", name, region, err))
			}

			// Ignore Default and Automatic EFS Backup Vaults in region (cannot be deleted)
			if name == "Default" || name == "aws/efs/automatic-backup-vault" {
				log.Printf("[INFO] Skipping Backup Vault: %s", name)
				continue
			}

			// Backup Vault deletion only supported when empty
			// Reference: https://docs.aws.amazon.com/aws-backup/latest/devguide/API_DeleteBackupVault.html
			if failedToDeleteRecoveryPoint {
				log.Printf("[INFO] Skipping Backup Vault (%s): not empty", name)
				continue
			}

			r := tfbackup.ResourceVault()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Backup Vaults sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Backup Vaults for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Backup Vaults for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccBackupVault_basic(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := sdkacctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
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

func TestAccBackupVault_withKMSKey(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := sdkacctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultWithKmsKey(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttrPair(resourceName, "kms_key_arn", "aws_kms_key.test", "arn"),
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

func TestAccBackupVault_withTags(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := sdkacctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.up", "down"),
					resource.TestCheckResourceAttr(resourceName, "tags.left", "right"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBackupVaultWithUpdateTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.up", "downdown"),
					resource.TestCheckResourceAttr(resourceName, "tags.left", "rightright"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
			{
				Config: testAccBackupVaultWithRemoveTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.fizz", "buzz"),
				),
			},
		},
	})
}

func TestAccBackupVault_disappears(t *testing.T) {
	var vault backup.DescribeBackupVaultOutput

	rInt := sdkacctest.RandInt()
	resourceName := "aws_backup_vault.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultExists(resourceName, &vault),
					acctest.CheckResourceDisappears(acctest.Provider, tfbackup.ResourceVault(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVaultDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault" {
			continue
		}

		input := &backup.DescribeBackupVaultInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeBackupVault(input)

		if err == nil {
			if *resp.BackupVaultName == rs.Primary.ID {
				return fmt.Errorf("Vault '%s' was not deleted properly", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckVaultExists(name string, vault *backup.DescribeBackupVaultOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn
		params := &backup.DescribeBackupVaultInput{
			BackupVaultName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeBackupVault(params)
		if err != nil {
			return err
		}

		*vault = *resp

		return nil
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

	input := &backup.ListBackupVaultsInput{}

	_, err := conn.ListBackupVaults(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBackupVaultConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"
}
`, randInt)
}

func testAccBackupVaultWithKmsKey(randInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Test KMS Key for AWS Backup Vault"
  deletion_window_in_days = 10
}

resource "aws_backup_vault" "test" {
  name        = "tf_acc_test_backup_vault_%d"
  kms_key_arn = aws_kms_key.test.arn
}
`, randInt)
}

func testAccBackupVaultWithTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    up   = "down"
    left = "right"
  }
}
`, randInt)
}

func testAccBackupVaultWithUpdateTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    up   = "downdown"
    left = "rightright"
    foo  = "bar"
    fizz = "buzz"
  }
}
`, randInt)
}

func testAccBackupVaultWithRemoveTags(randInt int) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = "tf_acc_test_backup_vault_%d"

  tags = {
    foo  = "bar"
    fizz = "buzz"
  }
}
`, randInt)
}
