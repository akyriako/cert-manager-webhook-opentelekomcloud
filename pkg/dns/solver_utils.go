package dns

import (
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/recordsets"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/dns/v2/zones"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
	"math/rand"
	"strconv"
	"strings"
)

const (
	primaryDnsIpAddress   string = "100.125.4.25"
	secondaryDnsIpAddress string = "100.125.129.199"
	authUrl               string = "https://iam.%s.otc.t-systems.com:443/v3"
	txtRecordSetType      string = "TXT"
)

var (
	OpenTelekomCloudDnsServers = []string{primaryDnsIpAddress, secondaryDnsIpAddress}
)

func GetOpenTelekomCloudDnsServerAddress() string {
	idx := rand.Intn(len(OpenTelekomCloudDnsServers))
	dnsServerAddress := OpenTelekomCloudDnsServers[idx]

	slog.Debug(fmt.Sprintf("opentelekomcloud nameserver %s will be used", dnsServerAddress))
	return dnsServerAddress
}

func (s *OpenTelekomCloudDnsProviderSolver) setOpenTelekomCloudDnsServiceClient(ch *v1alpha1.ChallengeRequest) error {
	if s.dnsClient != nil {
		return nil
	}

	config, err := loadConfig(ch.Config)
	if err != nil {
		return errors.Wrap(err, "loading challenge-request config failed")
	}

	slog.Debug("loaded challenge-request config")

	inCluster := false
	aksk := &OpenTelekomCloudAkSk{}
	err = env.Parse(aksk)
	if err != nil {
		slog.Debug(fmt.Sprintf("no ak/sk pair found in env variables, falling back to kubernetes secrets"))
		inCluster = true
	}

	if inCluster {
		aksk, err = s.getOpenTelekomCloudAkSk(config, ch)
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

func (s *OpenTelekomCloudDnsProviderSolver) getOpenTelekomCloudAkSk(
	config OpenTelekomCloudDnsProviderConfig,
	ch *v1alpha1.ChallengeRequest,
) (*OpenTelekomCloudAkSk, error) {
	ak, err := s.getSecret(ch.ResourceNamespace, config.AccessKeySecretRef)
	if err != nil {
		return nil, err
	}

	sk, err := s.getSecret(ch.ResourceNamespace, config.SecretKeySecretRef)
	if err != nil {
		return nil, err
	}

	aksk := OpenTelekomCloudAkSk{
		AccessKey: ak,
		SecretKey: sk,
	}

	return &aksk, nil
}

func (s *OpenTelekomCloudDnsProviderSolver) getSecret(namespace string, secretKeyRef *corev1.SecretKeySelector) (string, error) {
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

func (s *OpenTelekomCloudDnsProviderSolver) getResolvedZone(ch *v1alpha1.ChallengeRequest) (*zones.Zone, error) {
	action := strings.ToLower(string(ch.Action))

	listOpts := zones.ListOpts{
		Name: ch.ResolvedZone,
	}

	allPages, err := zones.List(s.dnsClient, listOpts).AllPages()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%s up failed", action))
	}

	allZones, err := zones.ExtractZones(allPages)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%s up failed", strings.ToLower(string(ch.Action))))
	}

	if len(allZones) != 1 {
		return nil, fmt.Errorf("%s failed: found %v while expecting 1 for zone %s", action, len(allZones), ch.ResolvedZone)
	}

	return &allZones[0], nil
}

func (s *OpenTelekomCloudDnsProviderSolver) getTxtRecordSetsByZone(ch *v1alpha1.ChallengeRequest, zone *zones.Zone) ([]recordsets.RecordSet, error) {
	action := strings.ToLower(string(ch.Action))

	recordsetsListOpts := recordsets.ListOpts{
		Name: ch.ResolvedFQDN,
		Type: txtRecordSetType,
		Data: getQuotedString(ch.Key),
	}

	allRecordPages, err := recordsets.ListByZone(s.dnsClient, zone.ID, recordsetsListOpts).AllPages()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%s failed", action))
	}

	allRecordSets, err := recordsets.ExtractRecordSets(allRecordPages)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("%s failed", action))
	}

	return allRecordSets, nil
}

func getQuotedString(s string) string {
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return s
	} else {
		return strconv.Quote(s)
	}
}
