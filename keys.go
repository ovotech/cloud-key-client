package keys

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"golang.org/x/oauth2/google"
	gcpiam "google.golang.org/api/iam/v1"
)

//Key type
type Key struct {
	Age           float64
	Name          string
	ID            string
	Provider      string
	LifeRemaining float64
}

//Provider type
type Provider struct {
	Provider   string
	GcpProject string
}

const (
	gcpTimeFormat           = "2006-01-02T15:04:05Z"
	gcpServiceAccountPrefix = "serviceAccounts/"
	gcpServiceAccountSuffix = "@"
	gcpKeyPrefix            = "keys/"
	gcpKeySuffix            = ""
	gcpProviderString       = "gcp"
	awsProviderString       = "aws"
)

//Keys returns a generic key slice of potentially multiple provider keys
func Keys(providers []Provider) (keys []Key) {
	for _, providerRequest := range providers {
		switch providerRequest.Provider {
		case "gcp":
			keys = appendSlice(keys, gcpKeys(providerRequest.GcpProject))
		case "aws":
			keys = appendSlice(keys, awsKeys())
		default:
			panic("No valid providers specified. Must be gcp|aws")
		}
	}
	return
}

//gcpKeys returns a slice of generic keys with provider=gcp
func gcpKeys(gcpProject string) (keys []Key) {
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, gcpiam.CloudPlatformScope)
	check(err)
	service, err := gcpiam.New(client)
	check(err)
	for _, acc := range gcpServiceAccounts(gcpProject, *service) {
		for _, gcpKey := range gcpServiceAccountKeys(gcpProject, acc.Email, *service) {
			//only iterate over keys that have a mins-to-expiry, or are of an age,
			// above a specific threshold, to differentiate between GCP-managed
			// and User-managed keys:
			// https://cloud.google.com/iam/docs/understanding-service-accounts
			keyAge := minsSince(parseTime(gcpTimeFormat, gcpKey.ValidAfterTime))
			keyMinsToExpiry := minsSince(parseTime(gcpTimeFormat,
				gcpKey.ValidBeforeTime))
			keys = append(keys, Key{keyAge,
				subString(gcpKey.Name, gcpServiceAccountPrefix,
					gcpServiceAccountSuffix),
				subString(gcpKey.Name, gcpKeyPrefix, gcpKeySuffix),
				gcpProviderString, keyMinsToExpiry})
		}
	}
	return
}

//gcpServiceAccounts returns a slice of GCP ServiceAccounts
func gcpServiceAccounts(project string, service gcpiam.Service) (accs []*gcpiam.ServiceAccount) {
	res, err := service.Projects.ServiceAccounts.List(fmt.Sprintf("projects/%s",
		project)).
		Do()
	check(err)
	accs = res.Accounts
	return
}

//gcpServiceAccountKeys returns a slice of ServiceAccountKeys
func gcpServiceAccountKeys(project, email string, service gcpiam.Service) (keys []*gcpiam.ServiceAccountKey) {
	res, err := service.Projects.ServiceAccounts.Keys.
		List(fmt.Sprintf("projects/%s/serviceAccounts/%s", project, email)).
		Do()
	check(err)
	keys = res.Keys
	return
}

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
		fmt.Println(*user.UserId)
		result, err := svc.ListAccessKeys(&awsiam.ListAccessKeysInput{
			MaxItems: aws.Int64(5),
			UserName: aws.String(*user.UserName),
		})
		check(err)
		for _, awsKey := range result.AccessKeyMetadata {
			keys = append(keys,
				Key{minsSince(*awsKey.CreateDate),
					*awsKey.UserName, *awsKey.AccessKeyId, awsProviderString, 0})
		}
	}
	return
}

//appendSlice appends the 2nd slice to the 1st, and returns the resulting slice
func appendSlice(keys, keysToAdd []Key) []Key {
	for _, keyToAdd := range keysToAdd {
		keys = append(keys, keyToAdd)
	}
	return keys
}

//check panics if error is not nil
func check(e error) {
	if e != nil {
		panic(e.Error())
	}
}

//parseTime calls time.Parse with the timeFormat and timeString provided, and
// checks for an error
func parseTime(timeFormat, timeString string) (then time.Time) {
	then, err := time.Parse(timeFormat, timeString)
	check(err)
	return
}

//minsSinceCreation returns the number of minutes since the provided time.Time
func minsSince(then time.Time) (minsSinceCreation float64) {
	duration := time.Since(then)
	minsSinceCreation = duration.Minutes()
	return
}

//substring returns a non-inclusive substring between the provided start and
// end strings. If neither start or end strings exist, it panics. Specify empty
// string as the 'end' parameter to use the length of str as the end index
func subString(str string, start string, end string) (result string) {
	startIndex := strings.Index(str, start)
	if startIndex != -1 {
		startIndex += len(start)
		endIndex := len(str)
		if len(end) > 0 {
			endIndex = strings.Index(str, end)
			if endIndex == -1 {
				panic("string " + end + "not found in target: " + str)
			}
		}
		result = str[startIndex:endIndex]
	} else {
		panic("string " + start + "not found in target: " + str)
	}
	return
}
