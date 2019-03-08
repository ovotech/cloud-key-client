package keys

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	gcpiam "google.golang.org/api/iam/v1"
)

//GcpKey type
type GcpKey struct{}

//keys returns a slice of keys from any authorised accounts
func (g GcpKey) keys(project string) (keys []Key, err error) {
	if err = validateGcpProjectString(project); err != nil {
		return
	}
	var iamService *gcpiam.Service
	if iamService, err = gcpIamService(); err != nil {
		return
	}
	var gcpSAs []*gcpiam.ServiceAccount
	if gcpSAs, err = gcpServiceAccounts(project, *iamService); err != nil {
		return
	}
	for _, acc := range gcpSAs {
		var gcpSAKeys []*gcpiam.ServiceAccountKey
		if gcpSAKeys, err = gcpServiceAccountKeys(gcpServiceAccountName(project, acc.Email), *iamService); err != nil {
			return
		}
		for _, gcpKey := range gcpSAKeys {
			var key Key
			if key, err = keyFromGcpKey(gcpKey, project); err != nil {
				return
			}
			keys = append(keys, key)
		}
	}
	return
}

func keyFromGcpKey(gcpKey *gcpiam.ServiceAccountKey, project string) (key Key, err error) {
	var timeCreated time.Time
	if timeCreated, err = time.Parse(gcpTimeFormat, gcpKey.ValidAfterTime); err != nil {
		return
	}
	var expiryTime time.Time
	if expiryTime, err = time.Parse(gcpTimeFormat, gcpKey.ValidBeforeTime); err != nil {
		return
	}
	var keyID string
	if keyID, err = subString(gcpKey.Name, gcpKeyPrefix, gcpKeySuffix); err != nil {
		return
	}
	var serviceAccountName string
	if serviceAccountName, err = subString(gcpKey.Name, gcpServiceAccountPrefix,
		gcpServiceAccountSuffix); err != nil {
		return
	}
	var fullServiceAccountName string
	if fullServiceAccountName, err = subString(gcpKey.Name,
		gcpServiceAccountPrefix, "/keys/"); err != nil {
		return
	}
	key = Key{
		serviceAccountName,
		fullServiceAccountName,
		time.Since(timeCreated).Minutes(),
		keyID,
		math.Abs(time.Since(expiryTime).Minutes()),
		strings.Join([]string{serviceAccountName,
			keyID[len(keyID)-numIDValuesInName:]}, "_"),
		Provider{gcpProviderString, project},
	}
	return
}

//createKey creates a key in the provided account
func (g GcpKey) createKey(project, account string) (keyID, newKey string, err error) {
	if err = validateGcpProjectString(project); err != nil {
		return
	}
	var iamService *gcpiam.Service
	if iamService, err = gcpIamService(); err != nil {
		return
	}
	var key *gcpiam.ServiceAccountKey
	if key, err = iamService.Projects.ServiceAccounts.Keys.
		Create(gcpServiceAccountName(project, account),
			&gcpiam.CreateServiceAccountKeyRequest{}).
		Do(); err != nil {
		return
	}
	newKey = key.PrivateKeyData
	nameSplit := strings.Split(key.Name, "/")
	keyID = nameSplit[len(nameSplit)-1]
	return
}

//deleteKey deletes the specified key from the specified account
func (g GcpKey) deleteKey(project, account, keyID string) (err error) {
	if err = validateGcpProjectString(project); err != nil {
		return
	}
	var iamService *gcpiam.Service
	if iamService, err = gcpIamService(); err != nil {
		return
	}
	_, err = iamService.Projects.ServiceAccounts.Keys.
		Delete(gcpServiceAccountKeyName(project, account, keyID)).
		Do()
	return
}

//gcpClient returns a new GCP IAM client
func gcpIamService() (service *gcpiam.Service, err error) {
	ctx := context.Background()
	var client *http.Client
	if client, err = google.DefaultClient(ctx, gcpiam.CloudPlatformScope); err != nil {
		return
	}
	return gcpiam.New(client)
}

//gcpServiceAccounts returns a slice of GCP ServiceAccounts
func gcpServiceAccounts(project string, service gcpiam.Service) (accs []*gcpiam.ServiceAccount, err error) {
	var res *gcpiam.ListServiceAccountsResponse
	if res, err = service.Projects.ServiceAccounts.
		List(gcpProjectName(project)).
		Do(); err != nil {
		return
	}
	accs = res.Accounts
	return
}

//gcpServiceAccountKeys returns a slice of ServiceAccountKeys
func gcpServiceAccountKeys(name string, service gcpiam.Service) (keys []*gcpiam.ServiceAccountKey, err error) {
	var res *gcpiam.ListServiceAccountKeysResponse
	if res, err = service.Projects.ServiceAccounts.Keys.
		List(name).
		KeyTypes("USER_MANAGED").
		Do(); err != nil {
		return
	}
	keys = res.Keys
	return
}

//gcpProjectName returns a string of the format "projects/{PROJECT}"
func gcpProjectName(project string) string {
	return fmt.Sprintf("projects/%s", project)
}

//gcpServiceAccountName returns a string of the format:
//  "projects/{PROJECT}/serviceAccounts/{SA}"
func gcpServiceAccountName(project, sa string) string {
	return fmt.Sprintf("%s/serviceAccounts/%s", gcpProjectName(project), sa)
}

//gcpServiceAccountKeyName returns a string of the format:
//  "projects/{PROJECT}/serviceAccounts/{SA}/keys/{KEY}"
func gcpServiceAccountKeyName(project, sa, key string) string {
	return fmt.Sprintf("%s/keys/%s", gcpServiceAccountName(project, sa), key)
}

//validateGcpProjectString validates the GCP project string
func validateGcpProjectString(project string) (err error) {
	if len(project) == 0 {
		err = errors.New("GCP project string needs to be set")
	}
	return
}
