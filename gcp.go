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
			keys = append(keys, Key{keyAge,
				strings.Join([]string{serviceAccountName,
					keyID[len(keyID)-numIDValuesInName:]}, "_"),
				keyID,
				gcpProviderString, keyMinsToExpiry})
		}
	}
	return
}

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

func gcpProjectName(project string) (name string) {
	name = fmt.Sprintf("projects/%s", project)
	return
}

func gcpServiceAccountName(project, email string) (name string) {
	name = fmt.Sprintf("projects/%s/serviceAccounts/%s", project, email)
	return
}

func createKey(project, sa string) (privateKeyData string) {
	csakr := gcpiam.CreateServiceAccountKeyRequest{}
	name := gcpServiceAccountName(project, sa)
	key, err := gcpClient().Projects.ServiceAccounts.Keys.
		Create(name, &csakr).
		Do()
	check(err)
	privateKeyData = key.PrivateKeyData
	return
}
