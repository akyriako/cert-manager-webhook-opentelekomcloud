# ACME webhook for Open Telekom Cloud DNS

Summary

## Installation

This webhook is installed exclusively via [Helm](https://helm.sh/). 

> [!NOTE]
> If you dont't have a Kubernetes cluster in place, this project
> comes with "batteries included"; a [Dev Container](https://containers.dev) (a `.devcontainer.json` file that can be found in the repo
> and will be discussed in a later chapter) that instructs any IDE that supports Dev Containers, to set up an isolated 
> containerized Kubernetes environment for you along with all necessary tooling (cert-manager, Helm etc.)

### Configuration 

Configure the Chart by setting the following parameters:

- `groupName`: sets environment variable `GROUP_NAME`, defaults to `acme.opentelekomcloud.com`
- `debug`: sets environment variable `OS_DEBUG`, defaults to `false`. When `true` lowers `slog.LogLevel` to `LevelDebug`
- `credentialsSecretRef`: a reference to the Kubernetes `Secret` that will hold the OTC access & secret keys, defaults to `cert-manager-webhook-opentelekomcloud-creds`
- `opentelekomcloud.accessKey`: the access key in plain text, **not required**
- `opentelekomcloud.secretKey`: the secret key in plain text, **not required**

> [!NOTE]
> The remaining chart variables are, besides self-explanatory, the same that used already by cert-manager-webhook-example 

### One-step

If `opentelekomcloud.accessKey` and `opentelekomcloud.secretKey` are **both set**, the chart will **automatically**:

- create the `credentialsSecretRef` secret
- encode `opentelekomcloud.accessKey` and `opentelekomcloud.secretKey` in base64
- populate secret's `data` with the encoded values of `opentelekomcloud.accessKey` and `opentelekomcloud.secretKey`

```bash
helm repo add otcdnswebhook https://www.github.com/akyriako/cert-manager-webhook-opentelekomcloud/
helm repo update

helm upgrade --install cmw-otc deploy/cert-manager-webhook-opentelekomcloud \
  --set opentelekomcloud.accessKey=$OS_ACCESS_KEY \
  --set opentelekomcloud.secretKey=$OS_SECRET_KEY \
  --namespace cert-manager
```

or you can alternatively override the [values.yaml](deploy%2Fcert-manager-webhook-opentelekomcloud%2Fvalues.yaml) and
set there the parameters.

### Two-steps

If for any reason the **one-step** installation is not fit for your deployment pipeline, you can split the installation 
in two steps. 

First deploy the webhook:

```bash
helm repo add otcdnswebhook https://www.github.com/akyriako/cert-manager-webhook-opentelekomcloud/
helm repo update

helm upgrade --install cmw-otc deploy/cert-manager-webhook-opentelekomcloud \
  --namespace cert-manager
```

and then create and deploy a `Secret` manifest, that would match the name of `credentialsSecretRef` value:

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

## Development

### Dev Container

### Conformance Testing

```bash
$ OS_DEBUG=true OS_ACCESS_KEY={AccessKeyinBase64} OS_SECRET_KEY={SecretKeyinBase64} TEST_ZONE_NAME=example.com. make test
```
> [!NOTE]
> Fill in the values of `OS_ACCESS_KEY` and `OS_SECRET_KEY`. Replace `example.com.` with your own (sub)domain.
> Make sure not to forget the trailing `.` in the `TEST_ZONE_NAME` value.
