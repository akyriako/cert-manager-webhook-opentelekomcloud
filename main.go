package main

import (
	"context"
	"fmt"
	"github.com/akyriako/cert-manager-webhook-opentelekomcloud/pkg/dns"
	"github.com/caarlos0/env/v10"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"k8s.io/klog/v2"
)

type config struct {
	GroupName string `env:"GROUP_NAME" envDefault:"acme.opentelekomcloud.com"`
	Debug     bool   `env:"OS_DEBUG" envDefault:"false"`
}

const (
	// As defined in /usr/include/sysexits.h => #define EX_CONFIG 78
	// For more information on linux exit codes and their special meaning, refer to:
	// https://tldp.org/LDP/abs/html/exitcodes.html
	exitCodeConfigurationError int = 78
)

var (
	cfg config
)

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		klog.Errorf(fmt.Sprintf("parsing env variables failed. %s", err.Error()))
		klog.FlushAndExit(klog.ExitFlushTimeout, exitCodeConfigurationError)
	}
}

func main() {
	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(cfg.GroupName, dns.NewOpenTelekomCloudDnsProviderSolver(context.Background()))
}
