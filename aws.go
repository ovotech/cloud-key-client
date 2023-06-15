package keys

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

// API ref: https://docs.aws.amazon.com/sdk-for-go/api/service/iam/

//AwsKey type
type AwsKey struct{}

const (
	awsAccessKeyLimit = 2
	defaultRegion     = "us-east-1"
	maxKeys           = 5
	maxUsers          = 1000
)

//Keys returns a slice of keys from any authorised accounts
func (a AwsKey) Keys(project string, includeInactiveKeys bool, token string) (keys []Key, err error) {
	var svc *awsiam.IAM
	if svc, err = iamService(); err != nil {
		return
	}
	var userList []*awsiam.User
	if userList, err = awsUserList(*svc); err != nil {
		return
	}
	for _, user := range userList {
		var keyList []*awsiam.AccessKeyMetadata
		if keyList, err = awsKeyList(*user.UserName, *svc); err != nil {
			return
		}
		for _, awsKey := range keyList {
			if includeInactiveKeys || *awsKey.Status == "Active" {
				keyID := *awsKey.AccessKeyId
				keys = append(keys, Key{
					*awsKey.UserName,
					*awsKey.UserName,
					time.Since(*awsKey.CreateDate).Minutes(),
					keyID,
					0,
					strings.Join([]string{*awsKey.UserName,
						keyID[len(keyID)-numIDValuesInName:]}, "_"),
					Provider{Provider: awsProviderString, GcpProject: "", Token: ""},
					*awsKey.Status,
				})
			}
		}
	}
	return
}

//CreateKey creates a key in the provided account
func (a AwsKey) CreateKey(project, account, token string) (keyID, newKey string, err error) {
	var svc *awsiam.IAM
	if svc, err = iamService(); err != nil {
		return
	}
	var keyList []*awsiam.AccessKeyMetadata
	if keyList, err = awsKeyList(account, *svc); err != nil {
		return
	}
	keyNum := len(keyList)
	if keyNum >= awsAccessKeyLimit {
		err = fmt.Errorf("Number of Access Keys for user: %s is already at its limit (%d)",
			account, awsAccessKeyLimit)
		return
	}
	var key *awsiam.CreateAccessKeyOutput
	if key, err = svc.CreateAccessKey(&awsiam.CreateAccessKeyInput{
		UserName: aws.String(account),
	}); err != nil {
		return
	}
	accessKey := key.AccessKey
	keyID = *accessKey.AccessKeyId
	newKey = *accessKey.SecretAccessKey
	return
}

//DeleteKey deletes the specified key from the specified account
func (a AwsKey) DeleteKey(project, account, keyID, token string) (err error) {
	var svc *awsiam.IAM
	if svc, err = iamService(); err != nil {
		return
	}
	_, err = svc.DeleteAccessKey(&awsiam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(keyID),
		UserName:    aws.String(account),
	})
	return
}

//awsSession creates a new AWS SDK session
func awsSession() (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region: aws.String(defaultRegion)},
	)
}

func iamService() (iamService *awsiam.IAM, err error) {
	var awsSess *session.Session
	if awsSess, err = awsSession(); err != nil {
		return
	}
	iamService = awsiam.New(awsSess)
	return
}

//awsUserList obtains a slice of Users from the AWS IAM service
func awsUserList(iamService awsiam.IAM) (users []*awsiam.User, err error) {
	var userResult *awsiam.ListUsersOutput
	if userResult, err = iamService.ListUsers(&awsiam.ListUsersInput{
		MaxItems: aws.Int64(maxUsers),
	}); err != nil {
		return
	}
	users = userResult.Users
	return
}

//awsKeyList obtains a slice of accessKeyMetadata from the specified User's account
//using the AWS IAM service
func awsKeyList(username string, iamService awsiam.IAM) (accessKeyMetadata []*awsiam.AccessKeyMetadata, err error) {
	var result *awsiam.ListAccessKeysOutput
	if result, err = iamService.ListAccessKeys(&awsiam.ListAccessKeysInput{
		MaxItems: aws.Int64(maxKeys),
		UserName: aws.String(username),
	}); err != nil {
		return
	}
	accessKeyMetadata = result.AccessKeyMetadata
	return
}
