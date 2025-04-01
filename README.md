[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/golang-standards/project-layout)
[![Go](https://img.shields.io/badge/go-1.22.0-blue?logo=go)](https://golang.org/)
[![Helm](https://img.shields.io/badge/helm-3.12.3-blue?logo=helm)](https://helm.sh/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.30.2-blue?logo=kubernetes)](https://kubernetes.io/)
[![Cert Manager](https://img.shields.io/badge/cert--manager-1.15.2-blue?logo=cert-manager)](https://cert-manager.io/)

![Alt text](.github/IONOS.CLOUD.BLU.svg?raw=true)

# IONOS Cloud DNS Webhook for cert-manager

This webhook allows you to utilize IONOS Cloud as a DNS provider for performing DNS-01 challenges when using [cert-manager](https://cert-manager.io/docs/).

## Overview

Cert-manager is a powerful Kubernetes add-on that automates the management and issuance of TLS certificates from various issuing sources. The IONOS Cloud Webhook extends cert-manager's capabilities to manage DNS challenges using IONOS Cloud's DNS services.

## Features

- Simplified integration with IONOS Cloud for DNS-01 challenges
- Secure, automated DNS record management for certificate validation
- Seamless issuance and renewal of TLS certificates

## Prerequisites

Before proceeding, ensure you have the following:
- A Kubernetes cluster with cert-manager installed
- An IONOS Cloud account with Cloud DNS API access
- kubectl configured to access your Kubernetes cluster

## Usage

1. ***Initiation of IONOS Cloud Authentication Token Secret:***
    See [IONOS Cloud Token management](https://docs.ionos.com/cloud/set-up-ionos-cloud/management/token-management) how to get a token.

    ```bash
    kubectl create secret generic cert-manager-webhook-ionos-cloud \
      --namespace=cert-manager \
      --from-literal=auth-token=<IONOS CLOUD AUTH TOKEN>
    ```
   
2. ***Install the webhook server***
    ```bash
    helm repo add cert-manager-webhook-ionos-cloud https://ionos-cloud.github.io/cert-manager-webhook-ionos-cloud
    helm upgrade cert-manager-webhook-ionos-cloud \
    --namespace cert-manager \
    --install cert-manager-webhook-ionos-cloud/cert-manager-webhook-ionos-cloud
    ```

> [!IMPORTANT]
> Before engaging into DNS-01, cert-manager does a DNS pre-check (SOA and NS records). Depending on your environment, you may see a failure in the cert-manager logs with the following message: `error When querying the SOA record for the domain...`. To fix the issue, you need to add the following arguments to the cert-manager: `--dns01-recursive-nameservers-only`, `--dns01-recursive-nameservers=8.8.8.8:53,1.1.1.1:53`. For more details, check out the official documentation: [https://cert-manager.io/docs/configuration/acme/dns01/#setting-nameservers-for-dns01-self-check](https://cert-manager.io/docs/configuration/acme/dns01/#setting-nameservers-for-dns01-self-check)


3. ***Using a custom cert-manager namespace (optional)***:

By convention, cert-manager is deployed in a namespace named `cert-manager`. The chart assumes this default and uses this value to add privileges to the cert-manager service account to enable the creation of resources of type "ionos-cloud". If you are deploying the cert-manager chart in a different namespace, you need to use the `certManager.namespace` chart value to set the name of the namespace where cert-manager is deployed. (e.g using `--set certManager.namespace=custom_namespace`)

4. ***Configuration of ClusterIssuer/Issuer:***

The first step of using cert-manager is creating an Issuer or ClusterIssuer. 

```yaml

apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: example@example.com # Replace this with your email address
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - dns01:
        webhook:
          solverName: ionos-cloud
          groupName: acme.ionos.com
```

   
5. ***Check with a demonstration of Ingress Integration with Wildcard SSL/TLS Certificate Generation***
   Given the preceding configuration, it is possible to exploit the capabilities of the Issuer or ClusterIssuer to
   dynamically produce wildcard SSL/TLS certificates in the following manner:
    ```yaml
    apiVersion: cert-manager.io/v1
    kind: Certificate
    metadata:
      name: wildcard-example
      namespace: default
    spec:
      secretName: wildcard-example-tls
      issuerRef:
        name: letsencrypt-prod
        kind: ClusterIssuer
      commonName: '*.example.runs.ionos.cloud' # project must be the owner of this zone
      duration: 8760h0m0s
      dnsNames:
        - example.runs.ionos.cloud
        - '*.example.runs.ionos.cloud'
    ---
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: app-ingress
      namespace: default
      annotations:
        ingress.kubernetes.io/rewrite-target: /
    spec:
      ingressClassName: "nginx"
      rules:
        - host: "app.example.runs.ionos.cloud"
          http:
            paths:
              - path: /
                pathType: Prefix
                backend:
                  service:
                    name: webapp
                    port:
                      number: 80
      tls:
        - hosts:
            - "app.example.runs.ionos.cloud"
          secretName: wildcard-example-tls
    ```

## Contribute

## Verify the image resource integrity

All official webhooks provided by IONOS are signed using [Cosign](https://docs.sigstore.dev/cosign/overview/).
The Cosign public key can be found in the [cosign.pub](./cosign.pub) file.

Note: Due to the early development stage of the webhook, the image is not yet signed
by [sigstores transparency log](https://github.com/sigstore/rekor).

```shell
export RELEASE_VERSION=latest
cosign verify --insecure-ignore-tlog --key cosign.pub ghcr.io/ionos-cloud/cert-manager-webhook-ionos-cloud:$RELEASE_VERSION
```

### Development Workflow

Check out the [make targets](https://www.gnu.org/software/make/manual/make.html) for the development cycle:

```bash
make help
```

### Conformance tests

 DNS providers must run the DNS01 provider conformance testing suite, else they will have undetermined behaviour when used with cert-manager.

 The conformance tests are provided by the cert-manager test package: https://github.com/cert-manager/cert-manager/blob/master/test/acme/suite.go

 To run the conformance tests: `TEST_ZONE_NAME=test-zone.com IONOS_TOKEN=api-token make conformance-test`

 the following environment variables must be set:
 
 * TEST_ZONE_NAME: the zone for which DNS-01 will be performed
 * IONOS_TOKEN: the token for accessing IONOS DNS API

### Compatibility:

This extension is **tested** with the following cert-manager major versions: v1.15.x, v1.16.x, v1.17.x