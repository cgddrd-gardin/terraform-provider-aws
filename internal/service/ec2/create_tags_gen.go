// Code generated by generators/createtags/main.go; DO NOT EDIT.

package keyvaluetags

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const EventualConsistencyTimeout = 5 * time.Minute

// Ec2CreateTags creates ec2 service tags for new resources.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func Ec2CreateTags(conn *ec2.EC2, identifier string, tagsMap interface{}) error {
	tags := New(tagsMap)
	input := &ec2.CreateTagsInput{
		Resources: aws.StringSlice([]string{identifier}),
		Tags:      tags.IgnoreAws().Ec2Tags(),
	}

	_, err := tfresource.RetryWhenNotFound(EventualConsistencyTimeout, func() (interface{}, error) {
		output, err := conn.CreateTags(input)

		if tfawserr.ErrCodeContains(err, ".NotFound") {
			err = &resource.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return output, err
	})

	if err != nil {
		return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
	}

	return nil
}