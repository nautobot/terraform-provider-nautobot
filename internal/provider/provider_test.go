package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	nb "github.com/nautobot/go-nautobot/v3"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"nautobot": providerserver.NewProtocol6WithError(New("test")),
}

type blockingRoundTripper struct{}

func (blockingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	<-req.Context().Done()
	return nil, req.Context().Err()
}

type successfulStatusRoundTripper struct {
	sawDeadline bool
}

func (r *successfulStatusRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	_, r.sawDeadline = req.Context().Deadline()
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"nautobot-version":"3.0.2"}`)),
		Request:    req,
	}, nil
}

func testAccPreCheck(t *testing.T) {
}

func testAccProviderConfig() string {
	url := testURL
	token := testToken

	return `
provider "nautobot" {
  url   = "` + url + `/api"
  token = "` + token + `"
}
`
}

func TestCheckVersionCompatibilityTimeout(t *testing.T) {
	t.Parallel()

	config := nb.NewConfiguration()
	config.Servers[0].URL = "http://nautobot.invalid/api"
	config.HTTPClient = &http.Client{Transport: blockingRoundTripper{}}
	api := nb.NewAPIClient(config)

	const timeout = 20 * time.Millisecond
	provider := &nautobotProvider{version: "3.0.2"}
	start := time.Now()
	err := provider.checkVersionCompatibility(context.Background(), api, timeout)

	if err == nil {
		t.Fatal("expected status request timeout error")
	}
	if !strings.Contains(err.Error(), "Nautobot /api/status/ request timed out after 20ms") {
		t.Fatalf("unexpected error: %s", err)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("status request timeout took too long: %s", elapsed)
	}
}

func TestCheckVersionCompatibilityTimeoutDisabled(t *testing.T) {
	t.Parallel()

	transport := &successfulStatusRoundTripper{}
	config := nb.NewConfiguration()
	config.Servers[0].URL = "http://nautobot.invalid/api"
	config.HTTPClient = &http.Client{Transport: transport}
	api := nb.NewAPIClient(config)

	provider := &nautobotProvider{version: "3.0.2"}
	if err := provider.checkVersionCompatibility(context.Background(), api, 0); err != nil {
		t.Fatalf("unexpected version compatibility error: %s", err)
	}
	if transport.sawDeadline {
		t.Fatal("status request had a deadline when timeout was disabled")
	}
}
