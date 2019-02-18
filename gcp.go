package keys

import (
	"context"
	"fmt"
	"math"
	"strings"

	"golang.org/x/oauth2/google"
	gcpiam "google.golang.org/api/iam/v1"
)

//GcpKey type
type GcpKey struct{}

//keys returns a slice of keys from any authorised accounts
func (g GcpKey) keys(project string) (keys []Key) {
	service := gcpClient()
	for _, acc := range gcpServiceAccounts(project, *service) {
		for _, gcpKey := range gcpServiceAccountKeys(gcpServiceAccountName(project, acc.Email),
			*service) {
			keyAge := minsSince(parseTime(gcpTimeFormat, gcpKey.ValidAfterTime))
			keyID := subString(gcpKey.Name, gcpKeyPrefix, gcpKeySuffix)
			keyMinsToExpiry := math.Abs(minsSince(parseTime(gcpTimeFormat,
				gcpKey.ValidBeforeTime)))
			serviceAccountName := subString(gcpKey.Name, gcpServiceAccountPrefix,
				gcpServiceAccountSuffix)
			fullServiceAccountName := subString(gcpKey.Name, gcpServiceAccountPrefix, "/keys/")
			keys = append(keys, Key{
				serviceAccountName,
				fullServiceAccountName,
				keyAge,
				keyID,
				keyMinsToExpiry,
				strings.Join([]string{serviceAccountName,
					keyID[len(keyID)-numIDValuesInName:]}, "_"),
				Provider{gcpProviderString, project},
			})
		}
	}
	return
}

//createKey creates a key in the provided account
func (g GcpKey) createKey(project, account string) (keyID, newKey string, err error) {
	key, err := gcpClient().Projects.ServiceAccounts.Keys.
		Create(gcpServiceAccountName(project, account),
			&gcpiam.CreateServiceAccountKeyRequest{}).
		Do()
	if err == nil {
		newKey = key.PrivateKeyData
		nameSplit := strings.Split(key.Name, "/")
		keyID = nameSplit[len(nameSplit)-1]
	}
	return
}

//deleteKey deletes the specified key from the specified account
func (g GcpKey) deleteKey(project, account, keyID string) (err error) {
	_, err = gcpClient().Projects.ServiceAccounts.Keys.
		Delete(gcpServiceAccountKeyName(project, account, keyID)).
		Do()
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
