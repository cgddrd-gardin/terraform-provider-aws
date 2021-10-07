package waf_test

import (
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_waf_sql_injection_match_set", &resource.Sweeper{
		Name: "aws_waf_sql_injection_match_set",
		F:    sweepSQLInjectionMatchSet,
		Dependencies: []string{
			"aws_waf_rate_based_rule",
			"aws_waf_rule",
			"aws_waf_rule_group",
		},
	})
}

func sweepSQLInjectionMatchSet(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).WAFConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error
	var g multierror.Group
	var mutex = &sync.Mutex{}

	input := &waf.ListSqlInjectionMatchSetsInput{}

	err = tfwaf.ListSQLInjectionMatchSetsPages(conn, input, func(page *waf.ListSqlInjectionMatchSetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, sqlInjectionMatchSet := range page.SqlInjectionMatchSets {
			r := tfwaf.ResourceSQLInjectionMatchSet()
			d := r.Data(nil)

			id := aws.StringValue(sqlInjectionMatchSet.SqlInjectionMatchSetId)
			d.SetId(id)

			// read concurrently and gather errors
			g.Go(func() error {
				// Need to Read first to fill in sql_injection_match_tuples attribute
				err := r.Read(d, client)

				if err != nil {
					sweeperErr := fmt.Errorf("error reading WAF SQL Injection Match Set (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				// In case it was already deleted
				if d.Id() == "" {
					return nil
				}

				mutex.Lock()
				defer mutex.Unlock()
				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))

				return nil
			})
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing WAF SQL Injection Matches for %s: %w", region, err))
	}

	if err = g.Wait().ErrorOrNil(); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error concurrently reading WAF SQL Injection Matches: %w", err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping WAF SQL Injection Matches for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping WAF SQL Injection Match sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func TestAccAWSWafSqlInjectionMatchSet_basic(t *testing.T) {
	var v waf.SqlInjectionMatchSet
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSQLInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sql_injection_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "URL_DECODE",
					}),
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

func TestAccAWSWafSqlInjectionMatchSet_changeNameForceNew(t *testing.T) {
	var before, after waf.SqlInjectionMatchSet
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	rNameNew := sdkacctest.RandomWithPrefix("tf-acc-test-new")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSQLInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSQLInjectionMatchSetChangeNameConfig(rNameNew),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rNameNew),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSWafSqlInjectionMatchSet_disappears(t *testing.T) {
	var v waf.SqlInjectionMatchSet
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSQLInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &v),
					testAccCheckSQLInjectionMatchSetDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSWafSqlInjectionMatchSet_changeTuples(t *testing.T) {
	var before, after waf.SqlInjectionMatchSet
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSQLInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "sql_injection_match_tuples.*", map[string]string{
						"field_to_match.#":      "1",
						"field_to_match.0.data": "",
						"field_to_match.0.type": "QUERY_STRING",
						"text_transformation":   "URL_DECODE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccSQLInjectionMatchSetConfig_changeTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSWafSqlInjectionMatchSet_noTuples(t *testing.T) {
	var sqlSet waf.SqlInjectionMatchSet
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_sql_injection_match_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, waf.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSQLInjectionMatchSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSQLInjectionMatchSetConfig_noTuples(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSQLInjectionMatchSetExists(resourceName, &sqlSet),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sql_injection_match_tuples.#", "0"),
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

func testAccCheckSQLInjectionMatchSetDisappears(v *waf.SqlInjectionMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn

		wr := tfwaf.NewRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateSqlInjectionMatchSetInput{
				ChangeToken:            token,
				SqlInjectionMatchSetId: v.SqlInjectionMatchSetId,
			}

			for _, sqlInjectionMatchTuple := range v.SqlInjectionMatchTuples {
				sqlInjectionMatchTupleUpdate := &waf.SqlInjectionMatchSetUpdate{
					Action: aws.String(waf.ChangeActionDelete),
					SqlInjectionMatchTuple: &waf.SqlInjectionMatchTuple{
						FieldToMatch:       sqlInjectionMatchTuple.FieldToMatch,
						TextTransformation: sqlInjectionMatchTuple.TextTransformation,
					},
				}
				req.Updates = append(req.Updates, sqlInjectionMatchTupleUpdate)
			}
			return conn.UpdateSqlInjectionMatchSet(req)
		})
		if err != nil {
			return fmt.Errorf("Error updating SqlInjectionMatchSet: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteSqlInjectionMatchSetInput{
				ChangeToken:            token,
				SqlInjectionMatchSetId: v.SqlInjectionMatchSetId,
			}
			return conn.DeleteSqlInjectionMatchSet(opts)
		})
		if err != nil {
			return fmt.Errorf("Error deleting SqlInjectionMatchSet: %s", err)
		}
		return nil
	}
}

func testAccCheckSQLInjectionMatchSetExists(n string, v *waf.SqlInjectionMatchSet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF SqlInjectionMatchSet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetSqlInjectionMatchSet(&waf.GetSqlInjectionMatchSetInput{
			SqlInjectionMatchSetId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.SqlInjectionMatchSet.SqlInjectionMatchSetId == rs.Primary.ID {
			*v = *resp.SqlInjectionMatchSet
			return nil
		}

		return fmt.Errorf("WAF SqlInjectionMatchSet (%s) not found", rs.Primary.ID)
	}
}

func testAccCheckSQLInjectionMatchSetDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_sql_injection_match_set" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetSqlInjectionMatchSet(
			&waf.GetSqlInjectionMatchSetInput{
				SqlInjectionMatchSetId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.SqlInjectionMatchSet.SqlInjectionMatchSetId == rs.Primary.ID {
				return fmt.Errorf("WAF SqlInjectionMatchSet %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the SqlInjectionMatchSet is already destroyed
		if tfawserr.ErrMessageContains(err, waf.ErrCodeNonexistentItemException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccSQLInjectionMatchSetConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = "%s"

  sql_injection_match_tuples {
    text_transformation = "URL_DECODE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccSQLInjectionMatchSetChangeNameConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = "%s"

  sql_injection_match_tuples {
    text_transformation = "URL_DECODE"

    field_to_match {
      type = "QUERY_STRING"
    }
  }
}
`, name)
}

func testAccSQLInjectionMatchSetConfig_changeTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = "%s"

  sql_injection_match_tuples {
    text_transformation = "NONE"

    field_to_match {
      type = "METHOD"
    }
  }
}
`, name)
}

func testAccSQLInjectionMatchSetConfig_noTuples(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_sql_injection_match_set" "test" {
  name = "%s"
}
`, name)
}
