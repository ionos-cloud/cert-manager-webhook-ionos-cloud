FROM gcr.io/distroless/static-debian12:nonroot

COPY cert-manager-webhook-ionos-cloud /cert-manager-webhook-ionos-cloud

ENTRYPOINT ["/cert-manager-webhook-ionos-cloud"]
