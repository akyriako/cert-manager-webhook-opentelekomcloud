package main

import (
	"context"
	"fmt"
	"github.com/akyriako/cert-manager-webhook-opentelekomcloud/pkg/dns"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"

	acmetest "github.com/cert-manager/cert-manager/test/acme"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

func init() {
	opts := zap.Options{}
	logger := zap.New(zap.UseFlagOptions(&opts))
	logf.SetLogger(logger)
}

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.

	dnsIpAddress := dns.GetOpenTelekomCloudDnsServerAddress()

	fixture := acmetest.NewFixture(dns.NewOpenTelekomCloudDnsProviderSolver(context.Background()),
		acmetest.SetResolvedZone(zone),
		acmetest.SetAllowAmbientCredentials(false),
		acmetest.SetManifestPath("testdata/opentelekomcloud"),
		acmetest.SetDNSServer(fmt.Sprintf("%s:53", dnsIpAddress)),
		acmetest.SetStrict(true),
	)

	fixture.RunConformance(t)
}
