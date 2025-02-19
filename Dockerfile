FROM gcr.io/distroless/static-debian11:nonroot

COPY cert-manager-webhook-ionos-cloud /cert-manager-webhook-ionos-cloud

ENTRYPOINT ["/cert-manager-webhook-ionos-cloud"]
