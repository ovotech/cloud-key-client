package keys

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

//ProviderInterface type
type ProviderInterface interface {
	Keys(project string, includeInactiveKeys bool, token string) (keys []Key, err error)
	CreateKey(project, account string) (keyID, newKey string, err error)
	DeleteKey(project, account, keyID string) (err error)
}

//Key type
type Key struct {
	Account       string
	FullAccount   string
	Age           float64
	ID            string
	LifeRemaining float64
	Name          string
	Provider      Provider
	Status        string
}

//Provider type
type Provider struct {
	Provider   string
	GcpProject string
	Token      string
}

const (
	aivenProviderString     = "aiven"
	aivenTimeFormat         = "2006-01-02T15:04:05Z"
	awsProviderString       = "aws"
	gcpTimeFormat           = "2006-01-02T15:04:05Z"
	gcpServiceAccountPrefix = "serviceAccounts/"
	gcpServiceAccountSuffix = "@"
	gcpKeyPrefix            = "keys/"
	gcpKeySuffix            = ""
	gcpProviderString       = "gcp"
	numIDValuesInName       = 6
)

var providerMap = map[string]ProviderInterface{
	aivenProviderString: AivenKey{},
	awsProviderString:   AwsKey{},
	gcpProviderString:   GcpKey{},
}

var logger = stdoutLogger().Sugar()

//RegisterProvider informs the tool about a new cloud provider, in addition to AWS and GCP, and registers it under a unique key
func RegisterProvider(providerName string, provider ProviderInterface) {
	providerMap[providerName] = provider
}

//Keys returns a generic key slice of potentially multiple provider keys
func Keys(providers []Provider, includeInactiveKeys bool) (keys []Key, err error) {
	for _, providerRequest := range providers {
		var providerKeys []Key
		if providerKeys, err = providerMap[providerRequest.Provider].
			Keys(providerRequest.GcpProject, includeInactiveKeys, providerRequest.Token); err != nil {
			return
		}
		keys = appendSlice(keys, providerKeys)
	}
	return
}

//CreateKeyFromScratch creates a new key from just provider and account
//parameters (an existing key is not required)
func CreateKeyFromScratch(provider Provider, account string) (string, string, error) {
	return providerMap[provider.Provider].CreateKey(provider.GcpProject, account)
}

//CreateKey creates a new key using details of the provided key
func CreateKey(key Key) (string, string, error) {
	return CreateKeyFromScratch(key.Provider, key.FullAccount)
}

//DeleteKey deletes the specified key
func DeleteKey(key Key) error {
	return providerMap[key.Provider.Provider].DeleteKey(key.Provider.GcpProject, key.FullAccount, key.ID)
}

//appendSlice appends the 2nd slice to the 1st, and returns the resulting slice
func appendSlice(keys, keysToAdd []Key) []Key {
	for _, keyToAdd := range keysToAdd {
		keys = append(keys, keyToAdd)
	}
	return keys
}

//substring returns a non-inclusive substring between the provided start and
// end strings. Specify empty string as the 'end' parameter to use the length of
// str as the end index
func subString(str string, start string, end string) (result string, err error) {
	defer logger.Sync()
	startIndex := strings.Index(str, start)
	if startIndex != -1 {
		startIndex += len(start)
		endIndex := len(str)
		if len(end) > 0 {
			endIndex = strings.Index(str, end)
			if endIndex == -1 {
				err = fmt.Errorf("string %s not found in target: %s", end, str)
				return
			}
		}
		result = str[startIndex:endIndex]
	} else {
		err = fmt.Errorf("string %s not found in target: %s", start, str)
		return
	}
	return
}

//stdoutLogger creates a stdout logger
func stdoutLogger() (logger *zap.Logger) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stdout"}
	logger, _ = config.Build()
	return
}
