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

{{- define "harbor.database.host" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- template "harbor.fullname" . }}-database
  {{- else -}}
    {{- .Values.database.external.host -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.port" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- printf "%s" "5432" -}}
  {{- else -}}
    {{- .Values.database.external.port -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.username" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- printf "%s" "postgres" -}}
  {{- else -}}
    {{- .Values.database.external.username -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.password" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- .Values.database.internal.password | b64enc | quote -}}
  {{- else -}}
    {{- .Values.database.external.password | b64enc | quote -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.rawPassword" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- .Values.database.internal.password -}}
  {{- else -}}
    {{- .Values.database.external.password -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.coreDatabase" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- printf "%s" "registry" -}}
  {{- else -}}
    {{- .Values.database.external.coreDatabase -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.clairDatabase" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- printf "%s" "postgres" -}}
  {{- else -}}
    {{- .Values.database.external.clairDatabase -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.notaryServerDatabase" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- printf "%s" "notaryserver" -}}
  {{- else -}}
    {{- .Values.database.external.notaryServerDatabase -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.notarySignerDatabase" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- printf "%s" "notarysigner" -}}
  {{- else -}}
    {{- .Values.database.external.notarySignerDatabase -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.clair" -}}
postgres://{{ template "harbor.database.username" . }}:{{ template "harbor.database.rawPassword" . }}@{{ template "harbor.database.host" . }}:{{ template "harbor.database.port" . }}/{{ template "harbor.database.clairDatabase" . }}?sslmode=disable
{{- end -}}

{{- define "harbor.database.notaryServer" -}}
postgres://{{ template "harbor.database.username" . }}:{{ template "harbor.database.rawPassword" . }}@{{ template "harbor.database.host" . }}:{{ template "harbor.database.port" . }}/{{ template "harbor.database.notaryServerDatabase" . }}?sslmode=disable
{{- end -}}

{{- define "harbor.database.notarySigner" -}}
postgres://{{ template "harbor.database.username" . }}:{{ template "harbor.database.rawPassword" . }}@{{ template "harbor.database.host" . }}:{{ template "harbor.database.port" . }}/{{ template "harbor.database.notarySignerDatabase" . }}?sslmode=disable
{{- end -}}
