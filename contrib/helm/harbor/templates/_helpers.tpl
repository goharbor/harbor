{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "harbor.name" -}}
{{- default "harbor" .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "harbor.fullname" -}}
{{- $name := default "harbor" .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Helm required labels */}}
{{- define "harbor.labels" -}}
heritage: {{ .Release.Service }}
release: {{ .Release.Name }}
chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app: "{{ template "harbor.name" . }}"
{{- end -}}

{{/* matchLabels */}}
{{- define "harbor.matchLabels" -}}
release: {{ .Release.Name }}
app: "{{ template "harbor.name" . }}"
{{- end -}}

{{- define "harbor.externalURL" -}}
{{- if .Values.externalPort -}}
{{- printf "%s:%s" .Values.externalDomain (toString .Values.externalPort) -}}
{{- else -}}
{{- .Values.externalDomain -}}
{{- end -}}
{{- end -}}

{{/*
Use *.domain.com as the Common Name in the certificate,
so it can match Harbor service FQDN and Notary service FQDN.
*/}}
{{- define "harbor.certCommonName" -}}
{{- $list := splitList "." .Values.externalDomain -}}
{{- $list := prepend (rest $list) "*" -}}
{{- $cn := join "." $list -}}
{{- printf "%s" $cn -}}
{{- end -}}

{{/* The external FQDN of Notary server. */}}
{{- define "harbor.notaryFQDN" -}}
{{- printf "notary-%s" .Values.externalDomain -}}
{{- end -}}

{{- define "harbor.notaryServiceName" -}}
{{- printf "%s-notary-server" (include "harbor.fullname" .) -}}
{{- end -}}
