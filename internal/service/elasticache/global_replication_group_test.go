package elasticache_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	resource.AddTestSweepers("aws_elasticache_global_replication_group", &resource.Sweeper{
		Name: "aws_elasticache_global_replication_group",
		F:    sweepGlobalReplicationGroups,
	})
}

// These timeouts are lower to fail faster during sweepers
const (
	sweeperGlobalReplicationGroupDisassociationReadyTimeout = 10 * time.Minute
	sweeperGlobalReplicationGroupDefaultUpdatedTimeout      = 10 * time.Minute
)

func sweepGlobalReplicationGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ElastiCacheConn

	var grgGroup multierror.Group

	input := &elasticache.DescribeGlobalReplicationGroupsInput{
		ShowMemberInfo: aws.Bool(true),
	}
	err = conn.DescribeGlobalReplicationGroupsPages(input, func(page *elasticache.DescribeGlobalReplicationGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, globalReplicationGroup := range page.GlobalReplicationGroups {
			globalReplicationGroup := globalReplicationGroup

			grgGroup.Go(func() error {
				id := aws.StringValue(globalReplicationGroup.GlobalReplicationGroupId)

				disassociationErrors := disassociateMembers(conn, globalReplicationGroup)
				if disassociationErrors != nil {
					sweeperErr := fmt.Errorf("failed to disassociate ElastiCache Global Replication Group (%s) members: %w", id, disassociationErrors)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}

				log.Printf("[INFO] Deleting ElastiCache Global Replication Group: %s", id)
				err := tfelasticache.DeleteGlobalReplicationGroup(conn, id, sweeperGlobalReplicationGroupDefaultUpdatedTimeout)
				if err != nil {
					sweeperErr := fmt.Errorf("error deleting ElastiCache Global Replication Group (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					return sweeperErr
				}
				return nil
			})
		}

		return !lastPage
	})

	grgErrs := grgGroup.Wait()

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping ElastiCache Global Replication Group sweep for %q: %s", region, err)
		return grgErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		grgErrs = multierror.Append(grgErrs, fmt.Errorf("error listing ElastiCache Global Replication Groups: %w", err))
	}

	return grgErrs.ErrorOrNil()
}

func disassociateMembers(conn *elasticache.ElastiCache, globalReplicationGroup *elasticache.GlobalReplicationGroup) error {
	var membersGroup multierror.Group

	for _, member := range globalReplicationGroup.Members {
		member := member

		if aws.StringValue(member.Role) == tfelasticache.GlobalReplicationGroupMemberRolePrimary {
			continue
		}

		id := aws.StringValue(globalReplicationGroup.GlobalReplicationGroupId)

		membersGroup.Go(func() error {
			if err := tfelasticache.DisassociateReplicationGroup(conn, id, aws.StringValue(member.ReplicationGroupId), aws.StringValue(member.ReplicationGroupRegion), sweeperGlobalReplicationGroupDisassociationReadyTimeout); err != nil {
				sweeperErr := fmt.Errorf(
					"error disassociating ElastiCache Replication Group (%s) in %s from Global Group (%s): %w",
					aws.StringValue(member.ReplicationGroupId), aws.StringValue(member.ReplicationGroupRegion), id, err,
				)
				log.Printf("[ERROR] %s", sweeperErr)
				return sweeperErr
			}
			return nil
		})
	}

	return membersGroup.Wait().ErrorOrNil()
}

func TestAccElastiCacheGlobalReplicationGroup_basic(t *testing.T) {
	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(primaryReplicationGroupResourceName, &primaryReplicationGroup),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "elasticache", regexp.MustCompile(`globalreplicationgroup:`+tfelasticache.GlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttrPair(resourceName, "at_rest_encryption_enabled", primaryReplicationGroupResourceName, "at_rest_encryption_enabled"),
					resource.TestCheckResourceAttr(resourceName, "auth_token_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "cache_node_type", primaryReplicationGroupResourceName, "node_type"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_enabled", primaryReplicationGroupResourceName, "cluster_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "engine", primaryReplicationGroupResourceName, "engine"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(resourceName, "actual_engine_version", primaryReplicationGroupResourceName, "engine_version"),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_id_suffix", rName),
					resource.TestMatchResourceAttr(resourceName, "global_replication_group_id", regexp.MustCompile(tfelasticache.GlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", tfelasticache.EmptyDescription),
					resource.TestCheckResourceAttr(resourceName, "primary_replication_group_id", primaryReplicationGroupId),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", "false"),
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

func TestAccElastiCacheGlobalReplicationGroup_description(t *testing.T) {
	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix("tf-acc-test")
	description1 := sdkacctest.RandString(10)
	description2 := sdkacctest.RandString(10)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", description1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", description2),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_disappears(t *testing.T) {
	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					acctest.CheckResourceDisappears(acctest.Provider, tfelasticache.ResourceGlobalReplicationGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_multipleSecondaries(t *testing.T) {
	var providers []*schema.Provider
	var globalReplcationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 3),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_MultipleSecondaries(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplcationGroup),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_ReplaceSecondary_differentRegion(t *testing.T) {
	var providers []*schema.Provider
	var globalReplcationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:        acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(&providers, 3),
		CheckDestroy:      testAccCheckReplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Setup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplcationGroup),
				),
			},
			{
				Config: testAccReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Move(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplcationGroup),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_clusterMode(t *testing.T) {
	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticache.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGlobalReplicationGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_ClusterMode(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
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

func testAccCheckGlobalReplicationGroupExists(resourceName string, v *elasticache.GlobalReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache Global Replication Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn
		grg, err := tfelasticache.FindGlobalReplicationGroupByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error retrieving ElastiCache Global Replication Group (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(grg.Status) == tfelasticache.GlobalReplicationGroupStatusDeleting || aws.StringValue(grg.Status) == tfelasticache.GlobalReplicationGroupStatusDeleted {
			return fmt.Errorf("ElastiCache Global Replication Group (%s) exists, but is in a non-available state: %s", rs.Primary.ID, aws.StringValue(grg.Status))
		}

		*v = *grg

		return nil
	}
}

func testAccCheckGlobalReplicationGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elasticache_global_replication_group" {
			continue
		}

		_, err := tfelasticache.FindGlobalReplicationGroupByID(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("ElastiCache Global Replication Group (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccPreCheckGlobalReplicationGroup(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn

	input := &elasticache.DescribeGlobalReplicationGroupsInput{}
	_, err := conn.DescribeGlobalReplicationGroups(input)

	if acctest.PreCheckSkipError(err) ||
		tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "Access Denied to API Version: APIGlobalDatastore") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
}
`, rName, primaryReplicationGroupId)
}

func testAccGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  global_replication_group_description = %[3]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, description)
}

func testAccGlobalReplicationGroupConfig_MultipleSecondaries(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", acctest.ProviderName, 1),
		testAccElasticacheVpcBaseWithProvider(rName, "alternate", acctest.ProviderNameAlternate, 1),
		testAccElasticacheVpcBaseWithProvider(rName, "third", acctest.ProviderNameThird, 1),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "alternate" {
  provider = awsalternate

  replication_group_id          = "%[1]s-a"
  replication_group_description = "alternate"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.alternate.name

  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "third" {
  provider = awsthird

  replication_group_id          = "%[1]s-t"
  replication_group_description = "third"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.third.name

  number_cache_clusters = 1
}
`, rName))
}

func testAccReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Setup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", acctest.ProviderName, 1),
		testAccElasticacheVpcBaseWithProvider(rName, "secondary", acctest.ProviderNameAlternate, 1),
		testAccElasticacheVpcBaseWithProvider(rName, "third", acctest.ProviderNameThird, 1),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "secondary" {
  provider = awsalternate

  replication_group_id          = "%[1]s-a"
  replication_group_description = "alternate"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.secondary.name

  number_cache_clusters = 1
}
`, rName))
}

func testAccReplicationGroupConfig_ReplaceSecondary_DifferentRegion_Move(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		testAccElasticacheVpcBaseWithProvider(rName, "primary", acctest.ProviderName, 1),
		testAccElasticacheVpcBaseWithProvider(rName, "secondary", acctest.ProviderNameAlternate, 1),
		testAccElasticacheVpcBaseWithProvider(rName, "third", acctest.ProviderNameThird, 1),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "third" {
  provider = awsthird

  replication_group_id          = "%[1]s-t"
  replication_group_description = "third"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.third.name

  number_cache_clusters = 1
}
`, rName))
}

func testAccGlobalReplicationGroupConfig_ClusterMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[1]q
  replication_group_description = "test"

  engine         = "redis"
  engine_version = "6.x"
  node_type      = "cache.m5.large"

  parameter_group_name       = "default.redis6.x.cluster.on"
  automatic_failover_enabled = true
  cluster_mode {
    num_node_groups         = 2
    replicas_per_node_group = 1
  }
}
`, rName)
}

func testAccElasticacheVpcBaseWithProvider(rName, name, provider string, subnetCount int) string {
	return acctest.ConfigCompose(
		testAccAvailableAZsNoOptInConfigWithProvider(name, provider),
		fmt.Sprintf(`
resource "aws_vpc" "%[1]s" {
  provider = %[2]s

  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "%[1]s" {
  provider = %[2]s

  count = %[4]d
	
  vpc_id            = aws_vpc.%[1]s.id
  cidr_block        = "192.168.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.%[1]s.names[count.index]
}

resource "aws_elasticache_subnet_group" "%[1]s" {
  provider = %[2]s
	
  name       = %[3]q
  subnet_ids = aws_subnet.%[1]s[*].id
}
`, name, provider, rName, subnetCount),
	)
}

func testAccAvailableAZsNoOptInConfigWithProvider(name, provider string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "%[1]s" {
  provider = %[2]s

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`, name, provider)
}
