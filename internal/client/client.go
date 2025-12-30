package client

import (
	"fmt"
	"time"

	"github.com/SSHcom/privx-sdk-go/v2/oauth"
	"github.com/SSHcom/privx-sdk-go/v2/restapi"
)

func authorize(apiBaseURL, bearerToken, apiClientID, apiClientSecret, oauthClientID, oauthClientSecret string) restapi.Authorizer {
	if bearerToken != "" {
		return oauth.WithToken("Bearer " + bearerToken)
	}

	// Create base client for OAuth
	baseClient := restapi.New(restapi.BaseURL(apiBaseURL))

	// Create OAuth authorizer with proper error handling
	auth := oauth.WithClientID(
		baseClient,
		oauth.Access(apiClientID),
		oauth.Secret(apiClientSecret),
		oauth.Digest(oauthClientID, oauthClientSecret),
	)

	return auth
}

func NewConnector(apiBaseURL, bearerToken, apiClientID, apiClientSecret, oauthClientID, oauthClientSecret string) (*restapi.Connector, error) {
	// Validate required parameters
	if apiBaseURL == "" {
		return nil, fmt.Errorf("API base URL is required")
	}

	if bearerToken == "" && (apiClientID == "" || apiClientSecret == "" || oauthClientID == "" || oauthClientSecret == "") {
		return nil, fmt.Errorf("either bearer token or complete OAuth credentials (client ID, client secret, OAuth client ID, OAuth client secret) are required")
	}

	auth := authorize(apiBaseURL, bearerToken, apiClientID, apiClientSecret, oauthClientID, oauthClientSecret)

	// Validate authentication before creating connector
	if auth == nil {
		return nil, fmt.Errorf("failed to create authentication object")
	}

	// Test authentication with retry logic
	var token string
	var err error
	for i := 0; i < 3; i++ {
		token, err = auth.AccessToken()
		if err == nil && token != "" {
			break
		}
		if i < 2 {
			// Brief pause before retry
			time.Sleep(100 * time.Millisecond)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("PrivX client authentication failed after 3 attempts: %v", err)
	}

	if token == "" {
		return nil, fmt.Errorf("authentication succeeded but received empty token")
	}

	connector := restapi.New(restapi.Auth(auth), restapi.Verbose(), restapi.BaseURL(apiBaseURL))
	return &connector, nil
}
