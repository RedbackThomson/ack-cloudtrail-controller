package trail

import (
	"context"

	ackcompare "github.com/aws-controllers-k8s/runtime/pkg/compare"
	ackrtlog "github.com/aws-controllers-k8s/runtime/pkg/runtime/log"
	svcsdk "github.com/aws/aws-sdk-go/service/cloudtrail"
)

// customUpdateTrail implements a custom logic for handling Trail
// resource updates.
func (rm *resourceManager) customUpdateTrail(
	ctx context.Context,
	desired *resource,
	latest *resource,
	delta *ackcompare.Delta,
) (*resource, error) {
	var err error
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.customUpdateTrail")
	defer func(err error) { exit(err) }(err)

	if delta.DifferentExcept("Spec.Tags") {
		err = rm.updateTrailField(ctx, desired)
		if err != nil {
			return nil, err
		}
	}
	if delta.DifferentAt("Spec.Tags") {
		err = rm.syncTrailTags(ctx, latest, desired)
		if err != nil {
			return nil, err
		}
	}
	readOneLatest, err := rm.ReadOne(ctx, desired)
	if err != nil {
		return nil, err
	}
	return rm.concreteResource(readOneLatest), nil
}

// syncTrailTags updates a trail list of tags.
func (rm *resourceManager) syncTrailTags(
	ctx context.Context,
	latest *resource,
	desired *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.syncTrailTags")
	defer func(err error) { exit(err) }(err)

	added, removed := computeTagsDelta(latest.ko.Spec.Tags, desired.ko.Spec.Tags)

	// Tags to create/update

	if len(removed) > 0 {
		_, err = rm.sdkapi.RemoveTagsWithContext(
			ctx,
			&svcsdk.RemoveTagsInput{
				ResourceId: (*string)(latest.ko.Status.ACKResourceMetadata.ARN),
				TagsList:   sdkTagsFromResourceTags(removed),
			},
		)
		rm.metrics.RecordAPICall("UPDATE", "RemoveTags", err)
		if err != nil {
			return err
		}
	}

	if len(added) > 0 {
		_, err = rm.sdkapi.AddTagsWithContext(
			ctx,
			&svcsdk.AddTagsInput{
				ResourceId: (*string)(latest.ko.Status.ACKResourceMetadata.ARN),
				TagsList:   sdkTagsFromResourceTags(added),
			},
		)
		rm.metrics.RecordAPICall("UPDATE", "AddTags", err)
		if err != nil {
			return err
		}
	}
	return nil
}

// updateTrail updates a given Trail fields.
func (rm *resourceManager) updateTrailField(
	ctx context.Context,
	desired *resource,
) (err error) {
	rlog := ackrtlog.FromContext(ctx)
	exit := rlog.Trace("rm.updateTrailField")
	defer func(err error) { exit(err) }(err)
	input := &svcsdk.UpdateTrailInput{
		Name: desired.ko.Spec.Name,
	}

	if desired.ko.Spec.CloudWatchLogsLogGroupARN != nil {
		input.SetCloudWatchLogsLogGroupArn(*desired.ko.Spec.CloudWatchLogsLogGroupARN)
	}
	if desired.ko.Spec.CloudWatchLogsRoleARN != nil {
		input.SetCloudWatchLogsRoleArn(*desired.ko.Spec.CloudWatchLogsRoleARN)
	}
	if desired.ko.Spec.EnableLogFileValidation != nil {
		input.SetEnableLogFileValidation(*desired.ko.Spec.EnableLogFileValidation)
	}
	if desired.ko.Spec.IncludeGlobalServiceEvents != nil {
		input.SetIncludeGlobalServiceEvents(*desired.ko.Spec.IncludeGlobalServiceEvents)
	}
	if desired.ko.Spec.IsMultiRegionTrail != nil {
		input.SetIsMultiRegionTrail(*desired.ko.Spec.IsMultiRegionTrail)
	}
	if desired.ko.Spec.IsOrganizationTrail != nil {
		input.SetIsOrganizationTrail(*desired.ko.Spec.IsOrganizationTrail)
	}
	if desired.ko.Spec.KMSKeyID != nil {
		input.SetKmsKeyId(*desired.ko.Spec.KMSKeyID)
	}
	if desired.ko.Spec.S3BucketName != nil {
		input.SetS3BucketName(*desired.ko.Spec.S3BucketName)
	}
	if desired.ko.Spec.S3KeyPrefix != nil {
		input.SetS3KeyPrefix(*desired.ko.Spec.S3KeyPrefix)
	}
	if desired.ko.Spec.SNSTopicName != nil {
		input.SetSnsTopicName(*desired.ko.Spec.SNSTopicName)
	}

	_, err = rm.sdkapi.UpdateTrailWithContext(ctx, input)
	rm.metrics.RecordAPICall("UPDATE", "UpdateTrail", err)
	return err
}
