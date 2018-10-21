package keys

import (
	"context"
	"fmt"
	"math"
	"strings"

	"golang.org/x/oauth2/google"
	gcpiam "google.golang.org/api/iam/v1"
)

//gcpKeys returns a slice of generic keys with provider=gcp
func gcpKeys(gcpProject string) (keys []Key) {
	service := gcpClient()
	for _, acc := range gcpServiceAccounts(gcpProject, *service) {
		for _, gcpKey := range gcpServiceAccountKeys(gcpServiceAccountName(gcpProject, acc.Email),
			*service) {
			keyAge := minsSince(parseTime(gcpTimeFormat, gcpKey.ValidAfterTime))
			keyID := subString(gcpKey.Name, gcpKeyPrefix, gcpKeySuffix)
			keyMinsToExpiry := math.Abs(minsSince(parseTime(gcpTimeFormat,
				gcpKey.ValidBeforeTime)))
			serviceAccountName := subString(gcpKey.Name, gcpServiceAccountPrefix,
				gcpServiceAccountSuffix)
			keys = append(keys, Key{
				serviceAccountName,
				keyAge,
				keyID,
				keyMinsToExpiry,
				strings.Join([]string{serviceAccountName,
					keyID[len(keyID)-numIDValuesInName:]}, "_"),
				Provider{gcpProviderString, gcpProject},
			})
		}
	}
	return
}

//gcpClient returns a new GCP IAM client
func gcpClient() (service *gcpiam.Service) {
	ctx := context.Background()
	var err error
	client, err := google.DefaultClient(ctx, gcpiam.CloudPlatformScope)
	check(err)
	service, err = gcpiam.New(client)
	check(err)
	return
}

//gcpServiceAccounts returns a slice of GCP ServiceAccounts
func gcpServiceAccounts(project string, service gcpiam.Service) (accs []*gcpiam.ServiceAccount) {
	res, err := service.Projects.ServiceAccounts.
		List(gcpProjectName(project)).
		Do()
	check(err)
	accs = res.Accounts
	return
}

//gcpServiceAccountKeys returns a slice of ServiceAccountKeys
func gcpServiceAccountKeys(name string, service gcpiam.Service) (keys []*gcpiam.ServiceAccountKey) {
	res, err := service.Projects.ServiceAccounts.Keys.
		List(name).
		KeyTypes("USER_MANAGED").
		Do()
	check(err)
	keys = res.Keys
	return
}

//gcpProjectName returns a string of the format "projects/{PROJECT}"
func gcpProjectName(project string) (name string) {
	name = fmt.Sprintf("projects/%s", project)
	return
}

//gcpServiceAccountName returns a string of the format:
//  "projects/{PROJECT}/serviceAccounts/{SA}"
func gcpServiceAccountName(project, sa string) (name string) {
	name = fmt.Sprintf("%s/serviceAccounts/%s", gcpProjectName(project), sa)
	return
}

//gcpServiceAccountKeyName returns a string of the format:
//  "projects/{PROJECT}/serviceAccounts/{SA}/keys/{KEY}"
func gcpServiceAccountKeyName(project, sa, key string) (name string) {
	name = fmt.Sprintf("%s/keys/%s", gcpServiceAccountName(project, sa), key)
	return
}

//GcpCreateKey creates a new service account key, returning the new key's
//private data if the creation was a success (nil if creation failed), and an
//error (nil upon success)
func gcpCreateKey(project, account string) (privateKeyData string, err error) {
	key, err := gcpClient().Projects.ServiceAccounts.Keys.
		Create(gcpServiceAccountName(project, account),
			&gcpiam.CreateServiceAccountKeyRequest{}).
		Do()
	if err != nil {
		privateKeyData = key.PrivateKeyData
	}
	return
}

//GcpDeleteKey deletes the specified service account key, and returns an error
//(nil upon successful deletion)
func gcpDeleteKey(project, account, keyID string) (err error) {
	_, err = gcpClient().Projects.ServiceAccounts.Keys.
		Delete(gcpServiceAccountKeyName(project, account, keyID)).
		Do()
	return
}
