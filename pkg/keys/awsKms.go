package keys

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	kms "github.com/aws/aws-sdk-go/service/kms"
)

// AwsKmsKey type
type AwsKmsKey struct{}

func (a AwsKmsKey) Keys(project string, includeInactiveKeys bool) (keys []Key, err error) {
	var svc *kms.KMS
	if svc, err = kmsService(); err != nil {
		return
	}
	// var kmsKeyList []*kms.ListKeysOutput
	output, err := svc.ListKeys(&kms.ListKeysInput{})
	for _, awsKmsKey := range output.Keys {
		// make call to describe key
		keyID := awsKmsKey.KeyId
		var describeOutput *kms.DescribeKeyOutput
		if describeOutput, err = svc.DescribeKey(
			&kms.DescribeKeyInput{KeyId: keyID}); err != nil {
			return
		}
		// describeOutput.
		// set key info in new key object
		keys = append(keys, Key{
			ID:  *keyID,
			Age: time.Since(*describeOutput.KeyMetadata.CreationDate).Minutes(),
		})
	}

	return
}

func (a AwsKmsKey) CreateKey(project, account string) (keyID, newKey string, err error) {

	return
}

func (a AwsKmsKey) DeleteKey(project, account, keyID string) (err error) {

	return
}

func kmsService() (kmsService *kms.KMS, err error) {
	var awsSess *session.Session
	if awsSess, err = awsSession(); err != nil {
		return
	}
	kmsService = kms.New(awsSess)
	return
}
