#!/bin/bash

# This script is for testing the IONOS cert-manager webhook end to end. 
# it deploys the chart on a Kubernetes cluster, creates an issuer,
# a certificate and tests the certificate is issued. 
#
# ARGUMENTS:
#   --cert-manager-version the cert-manager version.
# 
# Required environment variables: IONOS_TOKEN, TEST_ZONE_NAME

while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --cert-manager-version)
      cert_manager_version="$2"
      shift
      shift
      ;;
  esac
done


if [ -z "$cert_manager_version" ]; then
  echo "ERROR: --cert-manager-version flag is required!"
  exit 1
fi

if [ -z "$IONOS_TOKEN" ]; then
  echo "IONOS_TOKEN environment variable is required!"
  exit 1
fi


if [ -z "$TEST_ZONE_NAME" ]; then
  echo "TEST_ZONE_NAME environment variable is required!"
  exit 1
fi

# install cert-manager chart
 helm repo add jetstack https://charts.jetstack.io --force-update
            helm install \
            cert-manager jetstack/cert-manager \
            --namespace cert-manager \
            --create-namespace \
            --set crds.enabled=true \
             --set 'extraArgs={--dns01-recursive-nameservers-only,--dns01-recursive-nameservers=8.8.8.8:53\,1.1.1.1:53}' \
            --version $cert_manager_version


# build binary 
CGO_ENABLED=0 go build -ldflags "-s -w" -o ./cert-manager-webhook-ionos-cloud -v cmd/webhook/main.go

IMAGE_TAG=$(date +%N)

# build docker image
docker buildx build --platform linux/amd64 -t cert-manager-e2e-tests:$IMAGE_TAG .

# load image into kind
kind load docker-image cert-manager-e2e-tests:$IMAGE_TAG -n chart-testing

# install the chart
helm install cert-manager-webhook-ionos-cloud chart/cert-manager-webhook-ionos-cloud \
    --set image.repository=cert-manager-e2e-tests \
    --set image.tag=$IMAGE_TAG

# assert the deployment is ready
kubectl wait --timeout=30s --for=jsonpath='{.status.readyReplicas}'=1 deployment/cert-manager-webhook-ionos-cloud

# create the secret
kubectl create secret generic cert-manager-webhook-ionos-cloud --from-literal=auth-token="$(echo $IONOS_TOKEN)"

#create the issueq
kubectl apply -f .github/test-manifests/issuer.yaml

# assert the deployment is ready
kubectl wait --timeout=10s --for=condition=Ready=True issuer/letsencrypt-ionos-e2e

# create the test zone
PREFIX=$(date +%k%M%N)
ZONE_ID=$(curl -v -H "Authorization: Bearer $IONOS_TOKEN" --json \
"{\"properties\":{\"zoneName\":\"$PREFIX.$TEST_ZONE_NAME\",\"description\":\"used for e2e testing for cert-manager webhook\",\"enabled\":true}}" \
https://dns.de-fra.ionos.com/zones | jq -r .id)

# create certificate 
cat .github/test-manifests/certificate.tmpl.yaml | envsubst | kubectl apply -f -

# assert the certificate is ready
kubectl wait --timeout=2m --for=condition=Ready=True certificate/certificate-ionos-with-dns-01-test

curl -XDELETE -v -H "Authorization: Bearer $IONOS_TOKEN" \
"https://dns.de-fra.ionos.com/zones/$ZONE_ID"