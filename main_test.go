package main

import (
	"context"
	"encoding/json"
	"github.com/akyriako/cert-manager-webhook-opentelekomcloud/pkg/dns"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strings"
	"testing"

	acmetest "github.com/cert-manager/cert-manager/test/acme"
)

var (
	zone = os.Getenv("TEST_ZONE_NAME")
)

// HACK: Sets the controller-runtime logger otherwise test suite is raising a warning
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

func TestProjectNameConfiguration(t *testing.T) {
	cfg := dns.OpenTelekomCloudDnsProviderConfig{}
	if err := json.Unmarshal([]byte(`{"region":"eu-de"}`), &cfg); err != nil {
		t.Fatalf("decoding solver config failed: %v", err)
	}

	if cfg.ProjectName != "" {
		t.Fatalf("expected projectName to be unset, got %q", cfg.ProjectName)
	}

	encoded, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("encoding solver config failed: %v", err)
	}

	if strings.Contains(string(encoded), "projectName") {
		t.Fatalf("projectName must not be introduced in the config when not set, got %s", encoded)
	}

	cfg = dns.OpenTelekomCloudDnsProviderConfig{}
	if err := json.Unmarshal([]byte(`{"region":"eu-de","projectName":"my-project"}`), &cfg); err != nil {
		t.Fatalf("decoding solver config failed: %v", err)
	}

	if cfg.ProjectName != "my-project" {
		t.Fatalf("expected projectName to be my-project, got %q", cfg.ProjectName)
	}
}
