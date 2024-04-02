# ACME webhook for Open Telekom Cloud DNS

[Cert-manager](https://cert-manager.io/) DNS providers are integrations with various DNS (Domain Name System) service 
providers that allow cert-manager, a Kubernetes add-on, to automate the management of SSL/TLS certificates. 
DNS providers enable cert-manager to automatically perform challenges to prove domain ownership and obtain certificates 
from certificate authorities like [Let's Encrypt](https://letsencrypt.org/).

By configuring cert-manager with the compatible Open Telekom Cloud DNS provider, using this webhook, you can set up 
automatic certificate issuance and renewal for your Open Telekom Cloud CCE workloads without manual intervention. 
This automation is crucial for securing web applications and services deployed on CCE clusters.

## Installation

This webhook is installed exclusively via [Helm](https://helm.sh/). 

> [!NOTE]
> If you dont't have a Kubernetes cluster in place, this project
> comes with "batteries included"; a [Dev Container](https://containers.dev) (a `.devcontainer.json` file that can be 
> found in the repo and will be discussed in a later chapter) instructs any IDE that supports Dev Containers, to set up 
> an isolated containerized Kubernetes environment for you along with all necessary tooling (cert-manager, Helm etc.)

### Configuration 

Configure the Chart by setting the following parameters:

- `groupName`: sets environment variable `GROUP_NAME`, defaults to `acme.opentelekomcloud.com`
- `debug`: sets environment variable `OS_DEBUG`, defaults to `false`. When `true`, raises `klog` verbosity to `4`. It must be **boolean**
- `credentialsSecretRef`: a reference to the Kubernetes `Secret` that will hold the OTC access & secret keys, defaults to `cert-manager-webhook-opentelekomcloud-creds`
- `opentelekomcloud.accessKey`: the access key in plain text, **not required**
- `opentelekomcloud.secretKey`: the secret key in plain text, **not required**

> [!NOTE]
> The remaining chart variables are, besides self-explanatory, the same that used already by [cert-manager/webhook-example](https://github.com/cert-manager/webhook-example) 

### One-step

If `opentelekomcloud.accessKey` and `opentelekomcloud.secretKey` are **both set**, the chart will **automatically**:

- create the `credentialsSecretRef` secret
- encode `opentelekomcloud.accessKey` and `opentelekomcloud.secretKey` in base64
- populate secret's `data` with the encoded values of `opentelekomcloud.accessKey` and `opentelekomcloud.secretKey`

```bash
helm repo add cert-manager-webhook-opentelekomcloud https://www.github.com/akyriako/cert-manager-webhook-opentelekomcloud/
helm repo update

helm upgrade --install $CHART_RELEASE_NAME deploy/cert-manager-webhook-opentelekomcloud \
  --set opentelekomcloud.accessKey=$OS_ACCESS_KEY \
  --set opentelekomcloud.secretKey=$OS_SECRET_KEY \
  --namespace cert-manager
```

or you can alternatively override the [values.yaml](deploy%2Fcert-manager-webhook-opentelekomcloud%2Fvalues.yaml) and
set there the parameters.

### Two-steps

If for any reason the **one-step** installation is not fit for your deployment pipeline, you can split the installation 
in two steps:

First create and deploy a `Secret` manifest, that would match the name of `credentialsSecretRef` value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cert-manager-webhook-opentelekomcloud-creds
  namespace: cert-manager
type: Opaque
data:
  accessKey: "<ACCESS_KEY_in_Base64>"
  secretKey: "<SECRET_KEY_in_Base64>"
```

Deploy the secret with `kubectl`

and then deploy the webhook:

```bash
helm repo add cert-manager-webhook-opentelekomcloud https://www.github.com/akyriako/cert-manager-webhook-opentelekomcloud/
helm repo update

helm upgrade --install $CHART_RELEASE_NAME deploy/cert-manager-webhook-opentelekomcloud \
  --namespace cert-manager
```

## Usage

### Issuers & ClusterIssuers

`Issuers`, and `ClusterIssuers`, are Kubernetes resources that represent certificate authorities (CAs) that are able to 
generate signed certificates by honoring certificate signing requests. All cert-manager certificates require a 
referenced issuer that is in a ready condition to attempt to honor the request. The former is namespaced-scoped while the
latter is cluster-wide.

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: opentelekomcloud-letsencrypt-staging
  namespace: cert-manager
spec:
  acme:
    email: user@example.com
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: opentelekomcloud-letsencrypt-staging-tls-key
    solvers:
    - dns01:
        webhook:
          groupName: acme.opentelekomcloud.com
          solverName: opentelekomcloud
          config:
            region: "eu-de"
            accessKeySecretRef:
              name: cert-manager-webhook-opentelekomcloud-creds
              key: accessKey
            secretKeySecretRef:
              name: cert-manager-webhook-opentelekomcloud-creds
              key: secretKey
```

- `groupName` can be set in the respective chart parameter, otherwise defaults to `acme.opentelekomcloud.com`
- `solverName` should be `opentelekomcloud`, it is **not configurable**
- `region`, although configurable and required, it can only be set to `eu-de`
- `accessKeySecretRef` and `secretKeySecretRef` can be set in chart parameter `credentialsSecretRef`, if not defaults to `cert-manager-webhook-opentelekomcloud-creds`

Deploy the manifest above with `kubectl`.

### Certificate

In cert-manager, the `Certificate` resource represents a human readable definition of a certificate request. 
cert-manager uses this input to generate a private key and `CertificateRequest` resource in order to obtain a signed 
certificate from an `Issuer` or `ClusterIssuer`. The signed certificate and private key are then stored in the 
specified Secret resource. cert-manager will ensure that the certificate is auto-renewed before it expires and re-issued
if requested.

> [!IMPORTANT]
In order to issue any certificates, you'll need to configure an `Issuer` or `ClusterIssuer` resource first.

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: certificate-subdomain-example-com
  namespace: cert-manager
spec:
  dnsNames:
  - '*.subdomain.example.com'
  issuerRef:
    kind: ClusterIssuer
    name: opentelekomcloud-letsencrypt-staging
  secretName: certificate-subdomain-example-com-tls
```

Deploy the manifest above with `kubectl`.

## Development

If you dont't have a Kubernetes cluster in place, this project comes with "batteries included". A [Dev Container](https://containers.dev) 
(**.devcontainer/devcontainer.json**) will be used by any IDE that supports Dev Containers, to set up an isolated
containerized Kubernetes environment for you along with all necessary tooling (cert-manager, Helm etc.)

### Dev Container

> [!NOTE]
> Although you can use any IDE that supports Dev Containers, the extensions and features added on the base image are 
> tailored for Visual Studio Code.

#### Extensions & Features

A Dev Container will be created, with all the necessary prerequisites to get you started developing immediately. A
container, based on `mcr.microsoft.com/devcontainers/go:1.21-bullseye` will be spawned with the following features pre-installed:

- Golang 1.21
- Tooltitude for Go (Free License)
- Git, GitHub Actions, GitHub CLI, Git Graph
- Docker in Docker
- Kubectl, Helm, Helmfile, K9s, KinD, Dive
- [Bridge to Kubernetes](https://learn.microsoft.com/en-us/visualstudio/bridge/overview-bridge-to-kubernetes) Visual Studio Code Extension
- Resource Monitor

A `postCreateCommand` (**.devcontainer/setup.sh**) will provision:

- A containerized **Kubernetes cluster** with 1 control and 3 worker nodes **and** a private registry, using KinD (cluster manifest is in **.devcontainer/cluster.yaml**)
- A fully functional installation of Cert-Manager 

### Installation

In order to test the changes on a Kubernetes cluster, you need to build a new image, push the image to the container
registry of your choice and recreate the manifests that Helm will deploy to the Kubernetes cluster:

For building new image, execute:

```shell
make docker-build
```

For pushing the new image to a container registry, execute:

```shell
make docker-push
```

For creating the manifests out of the helm template, execute:

```shell
make rendered-manifest.yaml
```

> [!NOTE]
> Before executing the above target, you have to make sure that you have set the values of the following environment
> variables: `OS_ACCESS_KEY` and `OS_SECRET_KEY`.

The last one will create a yaml file that will contain all required manifests, `rendered-manifest.yaml`, in folder `_out`:
You can then deploy them in your Kubernetes cluster using `kubectl`:

```shell
kubectl apply -f _out/rendered-manifest.yaml
```

### Extend

#### Webhook Configuration

If you need to extend the webhook configuration via environment variables, you should extend struct `config`
which can be found in [main.go](main.go):

```go
type config struct {
	GroupName string `env:"GROUP_NAME" envDefault:"acme.opentelekomcloud.com"`
	Debug     bool   `env:"OS_DEBUG" envDefault:"false"`
}
```

> [!CAUTION]
> No sensitive information (either in plain or encoded text) should be added here for any reason.

Consequently, you might need to change the chart template values so they acknowledge and use the new parameters in the manifests.

#### DNS Solver Configuration

If you need to extend the configuration of the solver, extending its API Specs, you should extend struct `OpenTelekomCloudDnsProviderConfig`
which can be found in [config.go](pkg%2Fdns%2Fconfig.go):

```go
type OpenTelekomCloudDnsProviderConfig struct {
	// These fields will be set by users in the
	// `issuer.spec.acme.dns01.providers.webhook.config` field.
	Region             string                    `json:"region,required"`
	AccessKeySecretRef *corev1.SecretKeySelector `json:"accessKeySecretRef,omitempty"`
	SecretKeySecretRef *corev1.SecretKeySelector `json:"secretKeySecretRef,omitempty"`
}
```

> [!CAUTION]
> No sensitive information (either in plain or encoded text) should be added here for any reason.

Consequently, you might need to change the chart template values so they acknowledge and use the new parameters in the manifests.

#### Secrets

If you need to extend the secrets & credentials of the solver, you should extend struct `OpenTelekomCloudDnsProviderSecrets`
which can be found in [config.go](pkg%2Fdns%2Fconfig.go):

```go
type OpenTelekomCloudDnsProviderSecrets struct {
	AccessKey string `env:"OS_ACCESS_KEY,required"`
	SecretKey string `env:"OS_SECRET_KEY,required"`
}
```

Consequently, you might need to change the chart template values so they acknowledge and use the new parameters in the manifests.

> [!TIP]
> Access & Secret keys are enough to create an Open Telekom Cloud Provider Client and a DNS Service Client. User, Password,
> Domain or Tenant identifiers are not needed for the DNS Solver to work. 

### Conformance Testing

All DNS providers must run the DNS01 provider conformance testing suite, else they will have undetermined behaviour 
when used with cert-manager.

```bash
$ OS_DEBUG=true OS_ACCESS_KEY={AccessKeyinBase64} OS_SECRET_KEY={SecretKeyinBase64} TEST_ZONE_NAME=example.com. make test
```
> [!NOTE]
> Fill in the values of `OS_ACCESS_KEY` and `OS_SECRET_KEY`. Replace `example.com.` with your own (sub)domain.
> Make sure not to forget the trailing `.` in the `TEST_ZONE_NAME` value. You can omit any variable already defined
> in your session's environment variables.
