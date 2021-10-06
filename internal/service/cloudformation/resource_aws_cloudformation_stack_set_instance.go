package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/aws/internal/service/cloudformation"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/cloudformation/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/cloudformation/waiter"
	iamwaiter "github.com/hashicorp/terraform-provider-aws/aws/internal/service/iam/waiter"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfcloudformation "github.com/hashicorp/terraform-provider-aws/internal/service/cloudformation"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
)

func ResourceStackSetInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceStackSetInstanceCreate,
		Read:   resourceStackSetInstanceRead,
		Update: resourceStackSetInstanceUpdate,
		Delete: resourceStackSetInstanceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(tfcloudformation.StackSetInstanceCreatedDefaultTimeout),
			Update: schema.DefaultTimeout(tfcloudformation.StackSetInstanceUpdatedDefaultTimeout),
			Delete: schema.DefaultTimeout(tfcloudformation.StackSetInstanceDeletedDefaultTimeout),
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"parameter_overrides": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"retain_stack": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"stack_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stack_set_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceStackSetInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	region := meta.(*conns.AWSClient).Region
	if v, ok := d.GetOk("region"); ok {
		region = v.(string)
	}

	stackSetName := d.Get("stack_set_name").(string)
	input := &cloudformation.CreateStackInstancesInput{
		Accounts:     aws.StringSlice([]string{accountID}),
		Regions:      aws.StringSlice([]string{region}),
		StackSetName: aws.String(stackSetName),
	}

	if v, ok := d.GetOk("parameter_overrides"); ok {
		input.ParameterOverrides = expandParameters(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating CloudFormation StackSet Instance: %s", input)
	_, err := tfresource.RetryWhen(
		tfiam.PropagationTimeout,
		func() (interface{}, error) {
			input.OperationId = aws.String(resource.UniqueId())

			output, err := conn.CreateStackInstances(input)

			if err != nil {
				return nil, fmt.Errorf("error creating CloudFormation StackSet (%s) Instance: %w", stackSetName, err)
			}

			d.SetId(tfcloudformation.StackSetInstanceCreateResourceID(stackSetName, accountID, region))

			return tfcloudformation.WaitStackSetOperationSucceeded(conn, stackSetName, aws.StringValue(output.OperationId), d.Timeout(schema.TimeoutCreate))
		},
		func(err error) (bool, error) {
			if err == nil {
				return false, nil
			}

			message := err.Error()

			// IAM eventual consistency
			if strings.Contains(message, "AccountGate check failed") {
				return true, err
			}

			// IAM eventual consistency
			// User: XXX is not authorized to perform: cloudformation:CreateStack on resource: YYY
			if strings.Contains(message, "is not authorized") {
				return true, err
			}

			// IAM eventual consistency
			// XXX role has insufficient YYY permissions
			if strings.Contains(message, "role has insufficient") {
				return true, err
			}

			// IAM eventual consistency
			// Account XXX should have YYY role with trust relationship to Role ZZZ
			if strings.Contains(message, "role with trust relationship") {
				return true, err
			}

			// IAM eventual consistency
			if strings.Contains(message, "The security token included in the request is invalid") {
				return true, err
			}

			return false, fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) creation: %w", d.Id(), err)
		},
	)

	if err != nil {
		return err
	}

	return resourceStackSetInstanceRead(d, meta)
}

func resourceStackSetInstanceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	stackSetName, accountID, region, err := tfcloudformation.StackSetInstanceParseResourceID(d.Id())

	if err != nil {
		return err
	}

	stackInstance, err := tfcloudformation.FindStackInstanceByName(conn, stackSetName, accountID, region)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFormation StackSet Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudFormation StackSet Instance (%s): %w", d.Id(), err)
	}

	d.Set("account_id", stackInstance.Account)

	if err := d.Set("parameter_overrides", flattenAllCloudFormationParameters(stackInstance.ParameterOverrides)); err != nil {
		return fmt.Errorf("error setting parameters: %w", err)
	}

	d.Set("region", stackInstance.Region)
	d.Set("stack_id", stackInstance.StackId)
	d.Set("stack_set_name", stackSetName)

	return nil
}

func resourceStackSetInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	if d.HasChange("parameter_overrides") {
		stackSetName, accountID, region, err := tfcloudformation.StackSetInstanceParseResourceID(d.Id())

		if err != nil {
			return err
		}

		input := &cloudformation.UpdateStackInstancesInput{
			Accounts:           aws.StringSlice([]string{accountID}),
			OperationId:        aws.String(resource.UniqueId()),
			ParameterOverrides: []*cloudformation.Parameter{},
			Regions:            aws.StringSlice([]string{region}),
			StackSetName:       aws.String(stackSetName),
		}

		if v, ok := d.GetOk("parameter_overrides"); ok {
			input.ParameterOverrides = expandParameters(v.(map[string]interface{}))
		}

		log.Printf("[DEBUG] Updating CloudFormation StackSet Instance: %s", input)
		output, err := conn.UpdateStackInstances(input)

		if err != nil {
			return fmt.Errorf("error updating CloudFormation StackSet Instance (%s): %w", d.Id(), err)
		}

		if _, err := tfcloudformation.WaitStackSetOperationSucceeded(conn, stackSetName, aws.StringValue(output.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) update: %s", d.Id(), err)
		}
	}

	return resourceStackSetInstanceRead(d, meta)
}

func resourceStackSetInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFormationConn

	stackSetName, accountID, region, err := tfcloudformation.StackSetInstanceParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &cloudformation.DeleteStackInstancesInput{
		Accounts:     aws.StringSlice([]string{accountID}),
		OperationId:  aws.String(resource.UniqueId()),
		Regions:      aws.StringSlice([]string{region}),
		RetainStacks: aws.Bool(d.Get("retain_stack").(bool)),
		StackSetName: aws.String(stackSetName),
	}

	log.Printf("[DEBUG] Deleting CloudFormation StackSet Instance: %s", d.Id())
	output, err := conn.DeleteStackInstances(input)

	if tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeStackInstanceNotFoundException) || tfawserr.ErrCodeEquals(err, cloudformation.ErrCodeStackSetNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting CloudFormation StackSet Instance (%s): %s", d.Id(), err)
	}

	if _, err := tfcloudformation.WaitStackSetOperationSucceeded(conn, stackSetName, aws.StringValue(output.OperationId), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) deletion: %s", d.Id(), err)
	}

	return nil
}