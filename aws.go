package keys

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

const accessKeyLimit = 2

//awskeys returns a slice of generic keys with provider=aws
func awsKeys() (keys []Key) {
	// API ref: https://docs.aws.amazon.com/sdk-for-go/api/service/iam/
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

func awsSession() (sess *session.Session) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	check(err)
	return
}

func awsUserList(iamService awsiam.IAM) (users []*awsiam.User) {
	userResult, err := iamService.ListUsers(&awsiam.ListUsersInput{
		MaxItems: aws.Int64(10),
	})
	check(err)
	users = userResult.Users
	return
}

func awsKeyList(username string, iamService awsiam.IAM) (accessKeyMetadata []*awsiam.AccessKeyMetadata) {
	result, err := iamService.ListAccessKeys(&awsiam.ListAccessKeysInput{
		MaxItems: aws.Int64(5),
		UserName: aws.String(username),
	})
	check(err)
	accessKeyMetadata = result.AccessKeyMetadata
	return
}

func awsCreateKey(username string) (accessKeyID, secretAccessKey string) {
	//get number of keys that currently exist
	svc := awsiam.New(awsSession())
	if len(awsKeyList(username, *svc)) >= accessKeyLimit {
		panic("Number of Access Keys for user: " + username + "is already at its limit (" +
			strconv.Itoa(accessKeyLimit) +
			")")
	}
	key, err := svc.CreateAccessKey(&awsiam.CreateAccessKeyInput{
		UserName: aws.String(username),
	})
	check(err)
	accessKey := key.AccessKey
	accessKeyID = *accessKey.AccessKeyId
	secretAccessKey = *accessKey.SecretAccessKey
	return
}
