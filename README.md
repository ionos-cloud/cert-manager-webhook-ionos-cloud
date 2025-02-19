[![GoTemplate](https://img.shields.io/badge/go/template-black?logo=go)](https://github.com/golang-standards/project-layout)
[![Go](https://img.shields.io/badge/go-1.22.0-blue?logo=go)](https://golang.org/)
[![Helm](https://img.shields.io/badge/helm-3.12.3-blue?logo=helm)](https://helm.sh/)
[![Kubernetes](https://img.shields.io/badge/kubernetes-1.30.2-blue?logo=kubernetes)](https://kubernetes.io/)
[![Cert Manager](https://img.shields.io/badge/cert--manager-1.15.2-blue?logo=cert-manager)](https://cert-manager.io/)

![Alt text](.github/IONOS.CLOUD.BLU.svg?raw=true)

# IONOS Cloud DNS Webhook for cert-manager

Facilitate a webhook integration for leveraging the IONOS Cloud DNS alongside
its [API](https://ionos-cloud.github.io/rest-api/docs/dns/v1/) to act as a DNS01
ACME Issuer with [cert-manager](https://cert-manager.io/docs/).

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
    # [TODO] Add url to helm repository
    helm repo add cert-manager-webhook-ionos-cloud https://github.io/cert-manager-webhook-ionos-cloud/
    # [TODO] Add right parameters
    helm upgrade cert-manager-webhook-ionos-cloud --namespace cert-manager --install cert-manager-webhook-ionos-cloud/cert-manager-webhook-ionos-cloud --set IONOS_CLOUD_AUTH_TOKEN_SECRET_NAME=cert-manager-webhook-ionos-cloud
    ```

3. ***Configuration of ClusterIssuer/Issuer:***
   [TODO] document

   
4. ***Check with a demonstration of Ingress Integration with Wildcard SSL/TLS Certificate Generation***
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
        kind: Issuer
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
        kubernetes.io/ingress.class: "nginx"
    spec:
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

## Config Options

[TODO] document


## Contribute

### Development Workflow

Check out the [make targets](https://www.gnu.org/software/make/manual/make.html) for the development cycle:

```bash
make help
```

### Integration Tests


### Release Process Overview

