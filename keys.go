package keys

import (
	"strings"
	"time"
)

//Key type
type Key struct {
	Account       string
	FullAccount   string
	Age           float64
	ID            string
	LifeRemaining float64
	Name          string
	Provider      Provider
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
	numIDValuesInName       = 6
)

//Keys returns a generic key slice of potentially multiple provider keys
func Keys(providers []Provider) (keys []Key) {
	for _, providerRequest := range providers {
		switch providerRequest.Provider {
		case gcpProviderString:
			keys = appendSlice(keys, gcpKeys(providerRequest.GcpProject))
		case awsProviderString:
			keys = appendSlice(keys, awsKeys())
		default:
			panic("No valid providers specified. Must be gcp|aws")
		}
	}
	return
}

//CreateKeyFromScratch creates a new key from just provider and account
//parameters (an existing key is not required)
func CreateKeyFromScratch(provider Provider, account string) (keyID, newKey string, err error) {
	switch provider.Provider {
	case gcpProviderString:
		keyID, newKey, err = gcpCreateKey(provider.GcpProject, account)
	}
	return
}

//CreateKey creates a new key using details of the provided key
func CreateKey(key Key) (keyID, newKey string, err error) {
	switch key.Provider.Provider {
	case gcpProviderString:
		keyID, newKey, err = gcpCreateKey(key.Provider.GcpProject, key.FullAccount)
	}
	return
}

//DeleteKey deletes the specified key
func DeleteKey(key Key) (err error) {
	switch key.Provider.Provider {
	case gcpProviderString:
		err = gcpDeleteKey(key.Provider.GcpProject, key.FullAccount, key.ID)
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
