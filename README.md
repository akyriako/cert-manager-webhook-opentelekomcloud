# ACME webhook for Open Telekom Cloud DNS

Summary

## Installation

This webhook is installed exclusively via [Helm](https://helm.sh/). If you dont't have a Kubernetes cluster in place, this project
comes as well with "batteries included" via a Dev Container (a `.devcontainer.json` file that can be found in the repo
and will be discussed in a later chapter) that instructs any IDE that supports Dev Containers, to set up an isolated 
containerized Kubernetes environment for you along with all necessary tooling (cert-manager, Helm etc.)

### Configuration 

Configure the Chart by setting the following parameters:

- `groupName`: sets environment variable `GROUP_NAME`, defaults to `acme.opentelekomcloud.com`
- `debug`: sets environment variable `OS_DEBUG`, defaults to `false`. When `true` lowers `slog.LogLevel` to `LevelDebug`
- `credentialsSecretRef`: a reference to the Kubernetes `Secret` that will hold the OTC access & secret keys
- `opentelekomcloud.accessKey`: the access key in plain text, **not required**
- `opentelekomcloud.secretKey`: the secret key in plain text, **not required**

> [!NOTE]
> The remaining chart variables are, besides self-explanatory, the same that used already by cert-manager-webhook-example 

### One-step

If `opentelekomcloud.accessKey` and `opentelekomcloud.secretKey` are **both set**, the chart will **automatically**:

- create the `credentialsSecretRef` secret
- encode `opentelekomcloud.accessKey` and `opentelekomcloud.secretKey` in base64
- populate secret's `data` with the encoded values of `opentelekomcloud.accessKey` and `opentelekomcloud.secretKey`



### Two-steps

## Development

### Dev Container

### Conformance Testing

```bash
$ TEST_ZONE_NAME=example.com. make test
```

The example file has a number of areas you must fill in and replace with your
own options in order for tests to pass.
