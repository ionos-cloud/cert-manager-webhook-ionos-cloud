{{/*
Expand the name of the chart.
*/}}
{{- define "cert-manager-webhook-ionos-cloud.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "cert-manager-webhook-ionos-cloud.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cert-manager-webhook-ionos-cloud.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "cert-manager-webhook-ionos-cloud.labels" -}}
helm.sh/chart: {{ include "cert-manager-webhook-ionos-cloud.chart" . }}
{{ include "cert-manager-webhook-ionos-cloud.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "cert-manager-webhook-ionos-cloud.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cert-manager-webhook-ionos-cloud.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "cert-manager-webhook-ionos-cloud.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "cert-manager-webhook-ionos-cloud.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}


{{- define "cert-manager-webhook-ionos-cloud.selfSignedIssuer" -}}
{{ printf "%s-selfsign" (include "cert-manager-webhook-ionos-cloud.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-ionos-cloud.rootCAIssuer" -}}
{{ printf "%s-ca" (include "cert-manager-webhook-ionos-cloud.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-ionos-cloud.rootCACertificate" -}}
{{ printf "%s-ca" (include "cert-manager-webhook-ionos-cloud.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-ionos-cloud.servingCertificate" -}}
{{ printf "%s-webhook-tls" (include "cert-manager-webhook-ionos-cloud.fullname" .) }}
{{- end -}}
