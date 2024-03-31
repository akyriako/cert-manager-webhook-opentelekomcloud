package main

import (
	"context"
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

// Sets the controller-runtime logger otherwise test suite is raising a warning as
// it is not yet refactored to support newest structured logging logger (slog) that
// is introduced in Golang 1.21
func init() {
	opts := zap.Options{}
	logger := zap.New(zap.UseFlagOptions(&opts))
	logf.SetLogger(logger)
}

func TestRunsSuite(t *testing.T) {
	// The manifest path should contain a file named config.json that is a
	// snippet of valid configuration that should be included on the
	// ChallengeRequest passed as part of the test cases.

	//dnsIpAddress := dns.GetOpenTelekomCloudDnsServerAddress()

	fixture := acmetest.NewFixture(dns.NewOpenTelekomCloudDnsProviderSolver(context.Background()),
		acmetest.SetResolvedZone(zone),
		acmetest.SetAllowAmbientCredentials(false),
		acmetest.SetManifestPath("testdata/opentelekomcloud"),
		//acmetest.SetDNSServer(fmt.Sprintf("%s:53", dnsIpAddress)),
		//acmetest.SetDNSServer("8.8.8.8:53"),

		// Open Telekom Cloud DNS does not permit multiple TXT Records with the same name
		// in the same Record Set. The 'Present' challenge request in solver.go is updating
		// the Challenge Key value, if a TXT Record is with the same name is found in the Record Set.
		// For that reason, SetStrict option has to be set to 'false' in Open Telekom Cloud tests.
		// If set to 'true' the tests will **not** simulate creating and deleting multiple TXT Records
		// but updating the value of 'cert-manager-dns01-tests.example.com' with the new Challenge Key value.
		// That leads to skipping Extended/DeleteOneRetainsOthers test of the RunConformance suite.
		acmetest.SetStrict(false),
	)

	fixture.RunConformance(t)
}
