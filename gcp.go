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
	ctx := context.Background()
	client, err := google.DefaultClient(ctx, gcpiam.CloudPlatformScope)
	check(err)
	service, err := gcpiam.New(client)
	check(err)
	for _, acc := range gcpServiceAccounts(gcpProject, *service) {
		for _, gcpKey := range gcpServiceAccountKeys(gcpProject, acc.Email,
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
		KeyTypes("USER_MANAGED").
		Do()
	check(err)
	keys = res.Keys
	return
}
