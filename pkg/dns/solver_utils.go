package dns

import (
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
	"math/rand"
)

const (
	primaryDnsIpAddress   string = "100.125.4.25"
	secondaryDnsIpAddress string = "100.125.129.199"
	authUrl               string = "https://iam.%s.otc.t-systems.com:443/v3"
)

var (
	OpenTelekomCloudDnsServers = []string{primaryDnsIpAddress, secondaryDnsIpAddress}
)

func GetOpenTelekomCloudDnsServerAddress() string {
	idx := rand.Intn(2)
	ip := OpenTelekomCloudDnsServers[idx]

	slog.Debug(fmt.Sprintf("dns server %s will be used", ip))
	return ip
}

func (s *OpenTelekomCloudDnsProviderSolver) SetOpenTelekomCloudDnsServiceClient(ch *v1alpha1.ChallengeRequest) error {
	config, err := loadConfig(ch.Config)
	if err != nil {
		return errors.Wrap(err, "failed to load challenge-request config")
	}

	inCluster := false
	var aksk *OpenTelekomCloudAkSk
	err = env.Parse(aksk)
	if err != nil {
		slog.Debug(fmt.Sprintf("no ak/sk pair found in env variables, falling back to kubernetes secrets"))
		inCluster = true
	}

	if inCluster {
		aksk, err = s.GetOpenTelekomCloudAkSk(config, ch)
		if err != nil {
			return errors.Wrap(err, "failed to load access and secret keys")
		}
	}

	authOpts := golangsdk.AKSKAuthOptions{
		IdentityEndpoint: fmt.Sprintf(authUrl, config.Region),
		AccessKey:        aksk.AccessKey,
		SecretKey:        aksk.SecretKey,
	}

	endpointOpts := golangsdk.EndpointOpts{
		Region: config.Region,
	}

	providerClient, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return errors.Wrap(err, "creating an opentelekomcloud provider client failed")
	}

	dnsServiceClient, err := openstack.NewDNSV2(providerClient, endpointOpts)
	if err != nil {
		return errors.Wrap(err, "creating an opentelekomcloud dns service client failed")
	}

	slog.Debug("created an opentelekomcloud dns service client")

	s.dnsClient = dnsServiceClient
	return nil
}

func (s *OpenTelekomCloudDnsProviderSolver) GetOpenTelekomCloudAkSk(
	config OpenTelekomCloudDnsProviderConfig,
	ch *v1alpha1.ChallengeRequest,
) (*OpenTelekomCloudAkSk, error) {
	ak, err := s.GetSecret(ch.ResourceNamespace, config.AccessKeySecretRef)
	if err != nil {
		return nil, err
	}

	sk, err := s.GetSecret(ch.ResourceNamespace, config.SecretKeySecretRef)
	if err != nil {
		return nil, err
	}

	aksk := OpenTelekomCloudAkSk{
		AccessKey: ak,
		SecretKey: sk,
	}

	return &aksk, nil
}

func (s *OpenTelekomCloudDnsProviderSolver) GetSecret(namespace string, secretKeyRef *corev1.SecretKeySelector) (string, error) {
	secret, err := s.k8sClient.CoreV1().Secrets(namespace).Get(s.context, secretKeyRef.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("could not fetch secret %s: %w", secretKeyRef.Name, err)
	}

	data, ok := secret.Data[secretKeyRef.Key]
	if !ok {
		return "", fmt.Errorf("could not get key %s in secret %s", secretKeyRef.Key, secretKeyRef.Name)
	}

	slog.Debug(fmt.Sprintf("fetched secret: %s", secretKeyRef.Name))
	return string(data), nil
}
