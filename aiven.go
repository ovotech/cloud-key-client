package keys

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const aivenTokenEndpoint string = "https://api.aiven.io/v1/access_token"

// AivenKey type
type AivenKey struct{}

// Error type
type Error struct {
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}

// Token type
type Token struct {
	CreateTime      string `json:"create_time"`
	CurrentlyActive bool   `json:"currently_active"`
	Description     string `json:"description"`
	TokenPrefix     string `json:"token_prefix"`
}

// ListTokensResponse type
type ListTokensResponse struct {
	Errors  []Error `json:"errors"`
	Message string  `json:"message"`
	Tokens  []Token `json:"tokens"`
}

// CreateTokenResponse type
type CreateTokenResponse struct {
	CreateTime      string  `json:"create_time"`
	CreatedManually bool    `json:"created_manually"`
	Errors          []Error `json:"errors"`
	ExtendWhenUsed  bool    `json:"extend_when_used"`
	FullToken       string  `json:"full_token"`
	MaxAgeSeconds   int     `json:"max_age_seconds"`
	Message         string  `json:"message"`
	TokenPrefix     string  `json:"token_prefix"`
}

// RevokeTokenResponse type
type RevokeTokenResponse struct {
	Errors  []Error `json:"errors"`
	Message string  `json:"message"`
}

// Generic functions for sending an HTTP request
func doGenericHTTPReq(method, url, token string, payload io.Reader) (body []byte, err error) {
	client := http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// Get the listTokensResponse from the Aiven API
func listTokensResponse(token string) (ltr ListTokensResponse, err error) {
	// https://api.aiven.io/doc/#tag/User/operation/AccessTokenList
	body, err := doGenericHTTPReq(
		http.MethodGet,
		aivenTokenEndpoint,
		token,
		nil,
	)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &ltr)
	return
}

// Get the createTokenResponse from the Aiven API
func createTokenResponse(token, description string) (ctr CreateTokenResponse, err error) {
	// https://api.aiven.io/doc/#tag/User/operation/AccessTokenCreate
	jsonStr := []byte(fmt.Sprintf("{\"description\":\"%s\"}", description))
	body, err := doGenericHTTPReq(
		http.MethodPost,
		aivenTokenEndpoint,
		token,
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &ctr)
	return
}

// Get the revokeTokenResponse from the Aiven API
func revokeTokenResponse(tokenPrefix, token string) (rtr RevokeTokenResponse, err error) {
	// https://api.aiven.io/doc/#tag/User/operation/AccessTokenRevoke
	body, err := doGenericHTTPReq(
		http.MethodDelete,
		fmt.Sprintf("%s/%s", aivenTokenEndpoint, tokenPrefix),
		token,
		nil,
	)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &rtr)
	if err != nil {
		err = fmt.Errorf("Failed unmarshalling response: %s, response from Aiven API: %s", err, string(body[:]))
	}
	return
}

// Transform a slice of errors (returned in Aiven response) to a single error
func handleAPIErrors(errs []Error) (err error) {
	var errorMsgs []string
	for _, error := range errs {
		msg := fmt.Sprintf("msg: %s, status: %d", error.Message, error.Status)
		errorMsgs = append(errorMsgs, msg)
	}
	return errors.New(strings.Join(errorMsgs, ","))
}

// Return a status string (active|inactive)
func status(currentlyActive bool) string {
	status := "Inactive"
	if currentlyActive {
		status = "Active"
	}
	return status
}

// Get the description of a key/token from a 'fullAccount' identifier
func tokenDescriptionFromFullAccount(account string) string {
	splitAccount := strings.Split(account, "-")
	tokenPrefix := splitAccount[0]
	return account[len(tokenPrefix)+1:]
}

// Get the prefix of a key/token from a 'fullAccount' identifier
func tokenPrefixFromFullAccount(account string) string {
	splitAccount := strings.Split(account, "-")
	return splitAccount[0]
}

// Keys returns a slice of keys (or tokens in this case) for the user who
// owns the apiToken
func (a AivenKey) Keys(project string, includeInactiveKeys bool, apiToken string) (keys []Key, err error) {
	ltr, err := listTokensResponse(apiToken)
	if err != nil {
		return
	}
	if len(ltr.Errors) > 0 {
		err = handleAPIErrors(ltr.Errors)
		return
	}
	for _, token := range ltr.Tokens {
		var createTime time.Time
		if createTime, err = time.Parse(aivenTimeFormat, token.CreateTime); err != nil {
			return
		}
		// ignore the token if it has no description (this is the identifier
		// we use to track tokens down that are configured for rotation)
		if token.Description != "" {
			key := Key{
				FullAccount: fmt.Sprintf("%s-%s", token.TokenPrefix, token.Description),
				Age:         time.Since(createTime).Minutes(),
				ID:          token.TokenPrefix,
				Name:        token.Description,
				Provider:    Provider{Provider: aivenProviderString, Token: apiToken},
				Status:      status(token.CurrentlyActive),
			}
			keys = append(keys, key)
		}
	}
	return
}

// CreateKey creates a new Aiven API token
func (a AivenKey) CreateKey(project, account, token string) (keyID string, newKey string, err error) {
	description := tokenDescriptionFromFullAccount(account)
	ctr, err := createTokenResponse(token, description)
	if err != nil {
		return
	}
	if len(ctr.Errors) > 0 {
		err = handleAPIErrors(ctr.Errors)
		return
	}
	keyID = ctr.TokenPrefix
	newKey = ctr.FullToken
	return
}

// DeleteKey deletes the specified Aiven API token
func (a AivenKey) DeleteKey(project, account, keyID, token string) (err error) {
	tokenPrefix := tokenPrefixFromFullAccount(account)
	rtr, err := revokeTokenResponse(tokenPrefix, token)
	if err != nil {
		return
	}
	if len(rtr.Errors) > 0 {
		err = handleAPIErrors(rtr.Errors)
	}
	return
}
