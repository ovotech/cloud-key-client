package keys

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

//awskeys returns a slice of generic keys with provider=aws
func awsKeys() (keys []Key) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1")},
	)
	check(err)
	svc := awsiam.New(sess)
	userResult, err := svc.ListUsers(&awsiam.ListUsersInput{
		MaxItems: aws.Int64(10),
	})
	check(err)
	for _, user := range userResult.Users {
		result, err := svc.ListAccessKeys(&awsiam.ListAccessKeysInput{
			MaxItems: aws.Int64(5),
			UserName: aws.String(*user.UserName),
		})
		check(err)
		for _, awsKey := range result.AccessKeyMetadata {
			keyID := *awsKey.AccessKeyId
			keys = append(keys, Key{
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
