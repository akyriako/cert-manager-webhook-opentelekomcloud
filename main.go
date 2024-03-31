package main

import (
	"context"
	"fmt"
	"github.com/akyriako/cert-manager-webhook-opentelekomcloud/pkg/dns"
	"github.com/caarlos0/env/v10"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/cmd"
	"log/slog"
	"os"
)

type config struct {
	GroupName string `env:"GROUP_NAME" envDefault:"acme.opentelekomcloud.com"`
	Debug     bool   `env:"OS_DEBUG" envDefault:"false"`
}

const (
	exitCodeConfigurationError int = 78
)

var (
	cfg    config
	logger *slog.Logger
)

func init() {
	err := env.Parse(&cfg)
	if err != nil {
		slog.Error(fmt.Sprintf("parsing env variables failed. %s", err.Error()))
		os.Exit(exitCodeConfigurationError)
	}

	levelInfo := slog.LevelInfo
	if cfg.Debug {
		levelInfo = slog.LevelDebug
	}

	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: levelInfo,
	}))

	slog.SetDefault(logger)
}

func main() {
	// This will register our custom DNS provider with the webhook serving
	// library, making it available as an API under the provided GroupName.
	// You can register multiple DNS provider implementations with a single
	// webhook, where the Name() method will be used to disambiguate between
	// the different implementations.
	cmd.RunWebhookServer(cfg.GroupName, dns.NewOpenTelekomCloudDnsProviderSolver(context.Background()))
}
