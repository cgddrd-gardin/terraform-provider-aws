package ram

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/ram/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
	tfram "github.com/hashicorp/terraform-provider-aws/internal/service/ram"
)

const (
	ResourceShareInvitationStatusNotFound = "NotFound"
	ResourceShareInvitationStatusUnknown  = "Unknown"

	ResourceShareStatusNotFound = "NotFound"
	ResourceShareStatusUnknown  = "Unknown"

	PrincipalAssociationStatusNotFound = "NotFound"
)

// StatusResourceShareInvitation fetches the ResourceShareInvitation and its Status
func StatusResourceShareInvitation(conn *ram.RAM, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		invitation, err := tfram.FindResourceShareInvitationByARN(conn, arn)

		if err != nil {
			return nil, ResourceShareInvitationStatusUnknown, err
		}

		if invitation == nil {
			return nil, ResourceShareInvitationStatusNotFound, nil
		}

		return invitation, aws.StringValue(invitation.Status), nil
	}
}

// StatusResourceShareOwnerSelf fetches the ResourceShare and its Status
func StatusResourceShareOwnerSelf(conn *ram.RAM, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		share, err := tfram.FindResourceShareOwnerSelfByARN(conn, arn)

		if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
			return nil, ResourceShareStatusNotFound, nil
		}

		if err != nil {
			return nil, ResourceShareStatusUnknown, err
		}

		if share == nil {
			return nil, ResourceShareStatusNotFound, nil
		}

		return share, aws.StringValue(share.Status), nil
	}
}

func StatusResourceSharePrincipalAssociation(conn *ram.RAM, resourceShareArn, principal string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		association, err := tfram.FindResourceSharePrincipalAssociationByShareARNPrincipal(conn, resourceShareArn, principal)

		if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
			return nil, PrincipalAssociationStatusNotFound, err
		}

		if err != nil {
			return nil, ram.ResourceShareAssociationStatusFailed, err
		}

		if association == nil {
			return nil, ram.ResourceShareAssociationStatusDisassociated, nil
		}

		if aws.StringValue(association.Status) == ram.ResourceShareAssociationStatusFailed {
			extendedErr := fmt.Errorf("association status message: %s", aws.StringValue(association.StatusMessage))
			return association, aws.StringValue(association.Status), extendedErr
		}

		return association, aws.StringValue(association.Status), nil
	}
}