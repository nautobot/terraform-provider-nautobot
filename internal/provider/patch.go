package provider

import (
	"context"
	"fmt"
	"net/http"
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
