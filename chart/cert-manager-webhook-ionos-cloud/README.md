![Alt text](https://raw.githubusercontent.com/ionos-cloud/certbot-dns-ionos-cloud/main/.github/IONOS.CLOUD.BLU.svg)

# IONOS Cloud DNS Webhook for cert-manager

This webhook allows you to utilize IONOS Cloud as a DNS provider for performing DNS-01 challenges when using [cert-manager](https://cert-manager.io/docs/).

## Overview

Cert-manager is a powerful Kubernetes add-on that automates the management and issuance of TLS certificates from various issuing sources. The IONOS Cloud Webhook extends cert-manager's capabilities to manage DNS challenges using IONOS Cloud's DNS services.


## Usage

1. ***Initiation of IONOS Cloud Authentication Token Secret:***
    See [IONOS Cloud Token management](https://docs.ionos.com/cloud/set-up-ionos-cloud/management/token-management) for how to get a token.

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

3. ***Configuration of ClusterIssuer/Issuer:***

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

## Parameters

| Name        | Description           | Value  |
| ------------- |:-------------:| -----:|
| ionosCredentialsSecretName     | the name of the Kubernetes secret with the IONOS token     |  cert-manager-webhook-ionos-cloud |
| image.tag     | the container image tag name |   latest |
| image.pullPolicy     |  The image pull policy to be used for the container image    |   IfNotPresent |
| resources.limits.cpu      | The cpu limit for the container      |    |
| resources.limits.cpu      | The cpu limit for the container      |    |
| resources.limits.memory      | The memory limit for the container      |    |
| resources.requests.cpu      | The requested cpu for the container       |    |
| resources.requests.memory      | The requested memory for the container       |    |
| nodeSelector | The node selector for the pod |    {} |
| tolerations | Tolerations for the pod assignment    |    {}|
| affinity | Affinity for the pod     |    {} |
| service.port | The port exposed by the service     |    443 |
| service.type | The type of the service that exposes the pod      |    ClusterIP |