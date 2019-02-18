package keys

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

// API ref: https://docs.aws.amazon.com/sdk-for-go/api/service/iam/

//AwsKey type
type AwsKey struct{}

const accessKeyLimit = 2

//keys returns a slice of keys from any authorised accounts
func (a AwsKey) keys(project string) (keys []Key) {
	svc := awsiam.New(awsSession())
	for _, user := range awsUserList(*svc) {
		for _, awsKey := range awsKeyList(*user.UserName, *svc) {
			keyID := *awsKey.AccessKeyId
			keys = append(keys, Key{
				*awsKey.UserName,
				*awsKey.UserName,
				minsSince(*awsKey.CreateDate),
				keyID,
				0,
				strings.Join([]string{*awsKey.UserName,
					keyID[len(keyID)-numIDValuesInName:]}, "_"),
				Provider{awsProviderString, ""},
			})
		}
	}
	return
}

//createKey creates a key in the provided account
func (a AwsKey) createKey(project, account string) (keyID, newKey string, err error) {
	svc := awsiam.New(awsSession())
	if len(awsKeyList(account, *svc)) >= accessKeyLimit {
		panic("Number of Access Keys for user: " + account + "is already at its limit (" +
			strconv.Itoa(accessKeyLimit) +
			")")
	}
	key, err := svc.CreateAccessKey(&awsiam.CreateAccessKeyInput{
		UserName: aws.String(account),
	})
	accessKey := key.AccessKey
	keyID = *accessKey.AccessKeyId
	newKey = *accessKey.SecretAccessKey
	return
}

//deleteKey deletes the specified key from the specified account
func (a AwsKey) deleteKey(project, account, keyID string) (err error) {
	svc := awsiam.New(awsSession())
	_, err = svc.DeleteAccessKey(&awsiam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(keyID),
		UserName:    aws.String(account),
	})
	return
}

//awsSession creates a new AWS SDK session
func awsSession() (sess *session.Session) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	check(err)
	return
}

//awsUserList obtains a slice of Users from the AWS IAM service
func awsUserList(iamService awsiam.IAM) (users []*awsiam.User) {
	userResult, err := iamService.ListUsers(&awsiam.ListUsersInput{
		MaxItems: aws.Int64(10),
	})
	check(err)
	users = userResult.Users
	return
}

//awsKeyList obtains a slice of accessKeyMetadata from the specified User's account
//using the AWS IAM service
func awsKeyList(username string, iamService awsiam.IAM) (accessKeyMetadata []*awsiam.AccessKeyMetadata) {
	result, err := iamService.ListAccessKeys(&awsiam.ListAccessKeysInput{
		MaxItems: aws.Int64(5),
		UserName: aws.String(username),
	})
	check(err)
	accessKeyMetadata = result.AccessKeyMetadata
	return
}
