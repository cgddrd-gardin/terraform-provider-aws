package sagemaker_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsagemaker "github.com/hashicorp/terraform-provider-aws/internal/service/sagemaker"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_sagemaker_code_repository", &resource.Sweeper{
		Name: "aws_sagemaker_code_repository",
		F:    sweepCodeRepositories,
	})
}

func sweepCodeRepositories(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn

	err = conn.ListCodeRepositoriesPages(&sagemaker.ListCodeRepositoriesInput{}, func(page *sagemaker.ListCodeRepositoriesOutput, lastPage bool) bool {
		for _, instance := range page.CodeRepositorySummaryList {
			name := aws.StringValue(instance.CodeRepositoryName)

			input := &sagemaker.DeleteCodeRepositoryInput{
				CodeRepositoryName: instance.CodeRepositoryName,
			}

			log.Printf("[INFO] Deleting SageMaker Code Repository: %s", name)
			if _, err := conn.DeleteCodeRepository(input); err != nil {
				log.Printf("[ERROR] Error deleting SageMaker Code Repository (%s): %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Code Repository sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving SageMaker Code Repositorys: %w", err)
	}

	return nil
}

func TestAccAWSSagemakerCodeRepository_basic(t *testing.T) {
	var notebook sagemaker.DescribeCodeRepositoryOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCodeRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeRepositoryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "code_repository_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "git_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/hashicorp/terraform-provider-aws.git"),
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

func TestAccAWSSagemakerCodeRepository_gitConfig_branch(t *testing.T) {
	var notebook sagemaker.DescribeCodeRepositoryOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCodeRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeRepositoryGitBranchConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "code_repository_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "git_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.branch", "master"),
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

func TestAccAWSSagemakerCodeRepository_gitConfig_secret(t *testing.T) {
	var notebook sagemaker.DescribeCodeRepositoryOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCodeRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeRepositoryGitSecretConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "code_repository_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "git_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckResourceAttrPair(resourceName, "git_config.0.secret_arn", "aws_secretsmanager_secret.test", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCodeRepositoryGitSecretUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(resourceName, &notebook),
					resource.TestCheckResourceAttr(resourceName, "code_repository_name", rName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "sagemaker", fmt.Sprintf("code-repository/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "git_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "git_config.0.repository_url", "https://github.com/hashicorp/terraform-provider-aws.git"),
					resource.TestCheckResourceAttrPair(resourceName, "git_config.0.secret_arn", "aws_secretsmanager_secret.test2", "arn"),
				),
			},
		},
	})
}

func TestAccAWSSagemakerCodeRepository_disappears(t *testing.T) {
	var notebook sagemaker.DescribeCodeRepositoryOutput
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_sagemaker_code_repository.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCodeRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeRepositoryBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCodeRepositoryExists(resourceName, &notebook),
					acctest.CheckResourceDisappears(acctest.Provider, tfsagemaker.ResourceCodeRepository(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCodeRepositoryDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_code_repository" {
			continue
		}

		codeRepository, err := tfsagemaker.FindCodeRepositoryByName(conn, rs.Primary.ID)

		if tfawserr.ErrMessageContains(err, tfsagemaker.ErrCodeValidationException, "Cannot find CodeRepository") {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Sagemaker Code Repository (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(codeRepository.CodeRepositoryName) == rs.Primary.ID {
			return fmt.Errorf("sagemaker Code Repository %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCodeRepositoryExists(n string, codeRepo *sagemaker.DescribeCodeRepositoryOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Code Repository ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		resp, err := tfsagemaker.FindCodeRepositoryByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*codeRepo = *resp

		return nil
	}
}

func testAccCodeRepositoryBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
  }
}
`, rName)
}

func testAccCodeRepositoryGitBranchConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
    branch         = "master"
  }
}
`, rName)
}

func testAccCodeRepositoryGitSecretConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username = "example", password = "example" })
}

resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
    secret_arn     = aws_secretsmanager_secret.test.arn
  }

  depends_on = [aws_secretsmanager_secret_version.test]
}
`, rName)
}

func testAccCodeRepositoryGitSecretUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test2" {
  name = "%[1]s-2"
}

resource "aws_secretsmanager_secret_version" "test2" {
  secret_id     = aws_secretsmanager_secret.test2.id
  secret_string = jsonencode({ username = "example", password = "example" })
}

resource "aws_sagemaker_code_repository" "test" {
  code_repository_name = %[1]q

  git_config {
    repository_url = "https://github.com/hashicorp/terraform-provider-aws.git"
    secret_arn     = aws_secretsmanager_secret.test2.arn
  }

  depends_on = [aws_secretsmanager_secret_version.test2]
}
`, rName)
}
