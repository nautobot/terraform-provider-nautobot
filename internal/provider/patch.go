package provider

import (
	"context"
	"fmt"
	"net/http"

	nb "github.com/nautobot/go-nautobot/v2"
)

func NewSecurityProviderNautobotToken(t string) (*SecurityProviderNautobotToken, error) {
	return &SecurityProviderNautobotToken{
		token: t,
	}, nil
}

type SecurityProviderNautobotToken struct {
	token string
}

func (s *SecurityProviderNautobotToken) Intercept(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", s.token))
	return nil
}

func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int) *int32 {
	val := int32(i)
	return &val
}

func getStatusName(ctx context.Context, c *nb.APIClient, token string, statusID string) (string, error) {
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {
				Key:    token,
				Prefix: "Token",
			},
		},
	)

	// Fetch the status using the status ID
	status, _, err := c.ExtrasAPI.ExtrasStatusesRetrieve(auth, statusID).Execute()
	if err != nil {
		return "", err
	}

	// No need to dereference, just check if the string is empty
	if status.Name != "" {
		return status.Name, nil
	}

	return "", fmt.Errorf("status name not found for ID %s", statusID)
}

func getStatusID(ctx context.Context, c *nb.APIClient, token string, statusName string) (string, error) {
	auth := context.WithValue(
		ctx,
		nb.ContextAPIKeys,
		map[string]nb.APIKey{
			"tokenAuth": {
				Key:    token,
				Prefix: "Token",
			},
		},
	)

	statuses, _, err := c.ExtrasAPI.ExtrasStatusesList(auth).Name([]string{statusName}).Execute()
	if err != nil {
		return "", err
	}

	if len(statuses.Results) == 0 {
		return "", fmt.Errorf("status %s not found", statusName)
	}

	return statuses.Results[0].Id, nil
}
