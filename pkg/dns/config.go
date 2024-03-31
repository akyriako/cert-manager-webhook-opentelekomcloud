package dns

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	extapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// OpenTelekomCloudDnsProviderConfig is a structure that is used to decode into when
// solving a DNS01 challenge.
// This information is provided by cert-manager, and may be a reference to
// additional configuration that's needed to solve the challenge for this
// particular certificate or issuer.
// This typically includes references to Secret resources containing DNS
// provider credentials, in cases where a 'multi-tenant' DNS solver is being
// created.
// If you do *not* require per-issuer or per-certificate configuration to be
// provided to your webhook, you can skip decoding altogether in favour of
// using CLI flags or similar to provide configuration.
// You should not include sensitive information here. If credentials need to
// be used by your provider here, you should reference a Kubernetes Secret
// resource and fetch these credentials using a Kubernetes clientset.
type OpenTelekomCloudDnsProviderConfig struct {
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.
	Region             string                    `json:"region,required"`
	AccessKeySecretRef *corev1.SecretKeySelector `json:"accessKeySecretRef,omitempty"`
	SecretKeySecretRef *corev1.SecretKeySelector `json:"secretKeySecretRef,omitempty"`
}

// OpenTelekomCloudAkSk is a structure that is used to load the credentials from
// environment variables when solving a DNS01 challenge locally as the config.json
// refers only to Kubernetes secrets for the Open Telekom Cloud Access and Secret keys.
type OpenTelekomCloudAkSk struct {
	AccessKey string `env:"OS_ACCESS_KEY,required"`
	SecretKey string `env:"OS_SECRET_KEY,required"`
}

// loadConfig is a small helper function that decodes JSON configuration into
// the typed config struct.
func loadConfig(cfgJSON *extapi.JSON) (OpenTelekomCloudDnsProviderConfig, error) {
	cfg := OpenTelekomCloudDnsProviderConfig{}
	// handle the 'base case' where no configuration has been provided
	if cfgJSON == nil {
		return cfg, nil
	}
	if err := json.Unmarshal(cfgJSON.Raw, &cfg); err != nil {
		return cfg, fmt.Errorf("error decoding solver config: %v", err)
	}

	return cfg, nil
}
