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
If release name contains chart name it will be used as a full name.
*/}}
{{- define "harbor.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default "harbor" .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/* Helm required labels: legacy */}}
{{- define "harbor.legacy.labels" -}}
heritage: {{ .Release.Service }}
release: {{ .Release.Name }}
chart: {{ .Chart.Name }}
app: "{{ template "harbor.name" . }}"
{{- end -}}

{{/* Helm required labels */}}
{{- define "harbor.labels" -}}
heritage: {{ .Release.Service }}
release: {{ .Release.Name }}
chart: {{ .Chart.Name }}
app: "{{ template "harbor.name" . }}"
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: {{ include "harbor.name" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: {{ include "harbor.name" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
{{- end -}}

{{/* matchLabels */}}
{{- define "harbor.matchLabels" -}}
release: {{ .Release.Name }}
app: "{{ template "harbor.name" . }}"
{{- end -}}

{{/* Helper for printing values from existing secrets*/}}
{{- define "harbor.secretKeyHelper" -}}
  {{- if and (not (empty .data)) (hasKey .data .key) }}
    {{- index .data .key | b64dec -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.autoGenCert" -}}
  {{- if and .Values.expose.tls.enabled (eq .Values.expose.tls.certSource "auto") -}}
    {{- printf "true" -}}
  {{- else -}}
    {{- printf "false" -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.autoGenCertForIngress" -}}
  {{- if and (eq (include "harbor.autoGenCert" .) "true") (eq .Values.expose.type "ingress") -}}
    {{- printf "true" -}}
  {{- else -}}
    {{- printf "false" -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.autoGenCertForNginx" -}}
  {{- if and (eq (include "harbor.autoGenCert" .) "true") (ne .Values.expose.type "ingress") -}}
    {{- printf "true" -}}
  {{- else -}}
    {{- printf "false" -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.host" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- template "harbor.database" . }}
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

{{- define "harbor.database.rawPassword" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- $existingSecret := lookup "v1" "Secret" .Release.Namespace (include "harbor.database" .) -}}
    {{- if and (not (empty $existingSecret)) (hasKey $existingSecret.data "POSTGRES_PASSWORD") -}}
      {{- .Values.database.internal.password | default (index $existingSecret.data "POSTGRES_PASSWORD" | b64dec) -}}
    {{- else -}}
      {{- .Values.database.internal.password -}}
    {{- end -}}
  {{- else -}}
    {{- .Values.database.external.password -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.escapedRawPassword" -}}
  {{- include "harbor.database.rawPassword" . | urlquery | replace "+" "%20" -}}
{{- end -}}

{{- define "harbor.database.encryptedPassword" -}}
  {{- include "harbor.database.rawPassword" . | b64enc | quote -}}
{{- end -}}

{{- define "harbor.database.coreDatabase" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- printf "%s" "registry" -}}
  {{- else -}}
    {{- .Values.database.external.coreDatabase -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.database.sslmode" -}}
  {{- if eq .Values.database.type "internal" -}}
    {{- printf "%s" "disable" -}}
  {{- else -}}
    {{- .Values.database.external.sslmode -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.redis.scheme" -}}
  {{- with .Values.redis }}
    {{- ternary "redis+sentinel" "redis"  (and (eq .type "external" ) (not (not .external.sentinelMasterSet))) }}
  {{- end }}
{{- end -}}

/*host:port*/
{{- define "harbor.redis.addr" -}}
  {{- with .Values.redis }}
    {{- ternary (printf "%s:6379" (include "harbor.redis" $ )) .external.addr (eq .type "internal") }}
  {{- end }}
{{- end -}}

{{- define "harbor.redis.masterSet" -}}
  {{- with .Values.redis }}
    {{- ternary .external.sentinelMasterSet "" (eq "redis+sentinel" (include "harbor.redis.scheme" $)) }}
  {{- end }}
{{- end -}}

{{- define "harbor.redis.password" -}}
  {{- with .Values.redis }}
    {{- ternary "" .external.password (eq .type "internal") }}
  {{- end }}
{{- end -}}


{{- define "harbor.redis.pwdfromsecret" -}}
  {{- (lookup "v1" "Secret"  .Release.Namespace (.Values.redis.external.existingSecret)).data.REDIS_PASSWORD  | b64dec }}
{{- end -}}

{{- define "harbor.redis.cred" -}}
  {{- with .Values.redis }}
    {{- if (and (eq .type "external" ) (.external.existingSecret)) }}
      {{- printf ":%s@" (include "harbor.redis.pwdfromsecret" $) }}
    {{- else }}
      {{- ternary (printf "%s:%s@" (.external.username | urlquery) (.external.password | urlquery)) "" (and (eq .type "external" ) (not (not .external.password))) }}
    {{- end }}
  {{- end }}
{{- end -}}

/*scheme://[:password@]host:port[/master_set]*/
{{- define "harbor.redis.url" -}}
  {{- with .Values.redis }}
    {{- $path := ternary "" (printf "/%s" (include "harbor.redis.masterSet" $)) (not (include "harbor.redis.masterSet" $)) }}
    {{- printf "%s://%s%s%s" (include "harbor.redis.scheme" $) (include "harbor.redis.cred" $) (include "harbor.redis.addr" $) $path -}}
  {{- end }}
{{- end -}}

/*scheme://[:password@]addr/db_index?idle_timeout_seconds=30*/
{{- define "harbor.redis.urlForCore" -}}
  {{- with .Values.redis }}
    {{- $index := ternary "0" .external.coreDatabaseIndex (eq .type "internal") }}
    {{- printf "%s/%s?idle_timeout_seconds=30" (include "harbor.redis.url" $) $index -}}
  {{- end }}
{{- end -}}

/*scheme://[:password@]addr/db_index*/
{{- define "harbor.redis.urlForJobservice" -}}
  {{- with .Values.redis }}
    {{- $index := ternary .internal.jobserviceDatabaseIndex .external.jobserviceDatabaseIndex (eq .type "internal") }}
    {{- printf "%s/%s" (include "harbor.redis.url" $) $index -}}
  {{- end }}
{{- end -}}

/*scheme://[:password@]addr/db_index?idle_timeout_seconds=30*/
{{- define "harbor.redis.urlForRegistry" -}}
  {{- with .Values.redis }}
    {{- $index := ternary .internal.registryDatabaseIndex .external.registryDatabaseIndex (eq .type "internal") }}
    {{- printf "%s/%s?idle_timeout_seconds=30" (include "harbor.redis.url" $) $index -}}
  {{- end }}
{{- end -}}

/*scheme://[:password@]addr/db_index?idle_timeout_seconds=30*/
{{- define "harbor.redis.urlForTrivy" -}}
  {{- with .Values.redis }}
    {{- $index := ternary .internal.trivyAdapterIndex .external.trivyAdapterIndex (eq .type "internal") }}
    {{- printf "%s/%s?idle_timeout_seconds=30" (include "harbor.redis.url" $) $index -}}
  {{- end }}
{{- end -}}

/*scheme://[:password@]addr/db_index?idle_timeout_seconds=30*/
{{- define "harbor.redis.urlForHarbor" -}}
  {{- with .Values.redis }}
    {{- $index := ternary .internal.harborDatabaseIndex .external.harborDatabaseIndex (eq .type "internal") }}
    {{- printf "%s/%s?idle_timeout_seconds=30" (include "harbor.redis.url" $) $index -}}
  {{- end }}
{{- end -}}

/*scheme://[:password@]addr/db_index?idle_timeout_seconds=30*/
{{- define "harbor.redis.urlForCache" -}}
  {{- with .Values.redis }}
    {{- $index := ternary .internal.cacheLayerDatabaseIndex .external.cacheLayerDatabaseIndex (eq .type "internal") }}
    {{- printf "%s/%s?idle_timeout_seconds=30" (include "harbor.redis.url" $) $index -}}
  {{- end }}
{{- end -}}

{{- define "harbor.redis.dbForRegistry" -}}
  {{- with .Values.redis }}
    {{- ternary .internal.registryDatabaseIndex .external.registryDatabaseIndex (eq .type "internal") }}
  {{- end }}
{{- end -}}

{{- define "harbor.portal" -}}
  {{- printf "%s-portal" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.core" -}}
  {{- printf "%s-core" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.redis" -}}
  {{- printf "%s-redis" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.jobservice" -}}
  {{- printf "%s-jobservice" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.registry" -}}
  {{- printf "%s-registry" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.registryCtl" -}}
  {{- printf "%s-registryctl" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.database" -}}
  {{- printf "%s-database" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.trivy" -}}
  {{- printf "%s-trivy" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.nginx" -}}
  {{- printf "%s-nginx" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.exporter" -}}
  {{- printf "%s-exporter" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.ingress" -}}
  {{- printf "%s-ingress" (include "harbor.fullname" .) -}}
{{- end -}}

{{- define "harbor.noProxy" -}}
  {{- printf "%s,%s,%s,%s,%s,%s,%s,%s" (include "harbor.core" .) (include "harbor.jobservice" .) (include "harbor.database" .) (include "harbor.registry" .) (include "harbor.portal" .) (include "harbor.trivy" .) (include "harbor.exporter" .) .Values.proxy.noProxy -}}
{{- end -}}

{{- define "harbor.caBundleVolume" -}}
- name: ca-bundle-certs
  secret:
    secretName: {{ .Values.caBundleSecretName }}
{{- end -}}

{{- define "harbor.caBundleVolumeMount" -}}
- name: ca-bundle-certs
  mountPath: /harbor_cust_cert/custom-ca.crt
  subPath: ca.crt
{{- end -}}

{{/* scheme for all components because it only support http mode */}}
{{- define "harbor.component.scheme" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "https" -}}
  {{- else -}}
    {{- printf "http" -}}
  {{- end -}}
{{- end -}}

{{/* core component container port */}}
{{- define "harbor.core.containerPort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "8443" -}}
  {{- else -}}
    {{- printf "8080" -}}
  {{- end -}}
{{- end -}}

{{/* core component service port */}}
{{- define "harbor.core.servicePort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "443" -}}
  {{- else -}}
    {{- printf "80" -}}
  {{- end -}}
{{- end -}}

{{/* jobservice component container port */}}
{{- define "harbor.jobservice.containerPort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "8443" -}}
  {{- else -}}
    {{- printf "8080" -}}
  {{- end -}}
{{- end -}}

{{/* jobservice component service port */}}
{{- define "harbor.jobservice.servicePort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "443" -}}
  {{- else -}}
    {{- printf "80" -}}
  {{- end -}}
{{- end -}}

{{/* portal component container port */}}
{{- define "harbor.portal.containerPort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "8443" -}}
  {{- else -}}
    {{- printf "8080" -}}
  {{- end -}}
{{- end -}}

{{/* portal component service port */}}
{{- define "harbor.portal.servicePort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "443" -}}
  {{- else -}}
    {{- printf "80" -}}
  {{- end -}}
{{- end -}}

{{/* registry component container port */}}
{{- define "harbor.registry.containerPort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "5443" -}}
  {{- else -}}
    {{- printf "5000" -}}
  {{- end -}}
{{- end -}}

{{/* registry component service port */}}
{{- define "harbor.registry.servicePort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "5443" -}}
  {{- else -}}
    {{- printf "5000" -}}
  {{- end -}}
{{- end -}}

{{/* registryctl component container port */}}
{{- define "harbor.registryctl.containerPort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "8443" -}}
  {{- else -}}
    {{- printf "8080" -}}
  {{- end -}}
{{- end -}}

{{/* registryctl component service port */}}
{{- define "harbor.registryctl.servicePort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "8443" -}}
  {{- else -}}
    {{- printf "8080" -}}
  {{- end -}}
{{- end -}}

{{/* trivy component container port */}}
{{- define "harbor.trivy.containerPort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "8443" -}}
  {{- else -}}
    {{- printf "8080" -}}
  {{- end -}}
{{- end -}}

{{/* trivy component service port */}}
{{- define "harbor.trivy.servicePort" -}}
  {{- if .Values.internalTLS.enabled -}}
    {{- printf "8443" -}}
  {{- else -}}
    {{- printf "8080" -}}
  {{- end -}}
{{- end -}}

{{/* CORE_URL */}}
{{/* port is included in this url as a workaround for issue https://github.com/aquasecurity/harbor-scanner-trivy/issues/108 */}}
{{- define "harbor.coreURL" -}}
  {{- printf "%s://%s:%s" (include "harbor.component.scheme" .) (include "harbor.core" .) (include "harbor.core.servicePort" .) -}}
{{- end -}}

{{/* JOBSERVICE_URL */}}
{{- define "harbor.jobserviceURL" -}}
  {{- printf "%s://%s-jobservice" (include "harbor.component.scheme" .)  (include "harbor.fullname" .) -}}
{{- end -}}

{{/* PORTAL_URL */}}
{{- define "harbor.portalURL" -}}
  {{- printf "%s://%s" (include "harbor.component.scheme" .) (include "harbor.portal" .) -}}
{{- end -}}

{{/* REGISTRY_URL */}}
{{- define "harbor.registryURL" -}}
  {{- printf "%s://%s:%s" (include "harbor.component.scheme" .) (include "harbor.registry" .) (include "harbor.registry.servicePort" .) -}}
{{- end -}}

{{/* REGISTRY_CONTROLLER_URL */}}
{{- define "harbor.registryControllerURL" -}}
  {{- printf "%s://%s:%s" (include "harbor.component.scheme" .) (include "harbor.registry" .) (include "harbor.registryctl.servicePort" .) -}}
{{- end -}}

{{/* TOKEN_SERVICE_URL */}}
{{- define "harbor.tokenServiceURL" -}}
  {{- printf "%s/service/token" (include "harbor.coreURL" .) -}}
{{- end -}}

{{/* TRIVY_ADAPTER_URL */}}
{{- define "harbor.trivyAdapterURL" -}}
  {{- printf "%s://%s:%s" (include "harbor.component.scheme" .) (include "harbor.trivy" .) (include "harbor.trivy.servicePort" .) -}}
{{- end -}}

{{- define "harbor.internalTLS.core.secretName" -}}
  {{- if eq .Values.internalTLS.certSource "secret" -}}
    {{- .Values.internalTLS.core.secretName -}}
  {{- else -}}
    {{- printf "%s-core-internal-tls" (include "harbor.fullname" .) -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.internalTLS.jobservice.secretName" -}}
  {{- if eq .Values.internalTLS.certSource "secret" -}}
    {{- .Values.internalTLS.jobservice.secretName -}}
  {{- else -}}
    {{- printf "%s-jobservice-internal-tls" (include "harbor.fullname" .) -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.internalTLS.portal.secretName" -}}
  {{- if eq .Values.internalTLS.certSource "secret" -}}
    {{- .Values.internalTLS.portal.secretName -}}
  {{- else -}}
    {{- printf "%s-portal-internal-tls" (include "harbor.fullname" .) -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.internalTLS.registry.secretName" -}}
  {{- if eq .Values.internalTLS.certSource "secret" -}}
    {{- .Values.internalTLS.registry.secretName -}}
  {{- else -}}
    {{- printf "%s-registry-internal-tls" (include "harbor.fullname" .) -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.internalTLS.trivy.secretName" -}}
  {{- if eq .Values.internalTLS.certSource "secret" -}}
    {{- .Values.internalTLS.trivy.secretName -}}
  {{- else -}}
    {{- printf "%s-trivy-internal-tls" (include "harbor.fullname" .) -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.tlsCoreSecretForIngress" -}}
  {{- if eq .Values.expose.tls.certSource "none" -}}
    {{- printf "" -}}
  {{- else if eq .Values.expose.tls.certSource "secret" -}}
    {{- .Values.expose.tls.secret.secretName -}}
  {{- else -}}
    {{- include "harbor.ingress" . -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.tlsSecretForNginx" -}}
  {{- if eq .Values.expose.tls.certSource "secret" -}}
    {{- .Values.expose.tls.secret.secretName -}}
  {{- else -}}
    {{- include "harbor.nginx" . -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.metricsPortName" -}}
  {{- if .Values.internalTLS.enabled }}
    {{- printf "https-metrics" -}}
  {{- else -}}
    {{- printf "http-metrics" -}}
  {{- end -}}
{{- end -}}

{{- define "harbor.traceEnvs" -}}
  TRACE_ENABLED: "{{ .Values.trace.enabled }}"
  TRACE_SAMPLE_RATE: "{{ .Values.trace.sample_rate }}"
  TRACE_NAMESPACE: "{{ .Values.trace.namespace }}"
  {{- if .Values.trace.attributes }}
  TRACE_ATTRIBUTES: {{ .Values.trace.attributes | toJson | squote }}
  {{- end }}
  {{- if eq .Values.trace.provider "jaeger" }}
  TRACE_JAEGER_ENDPOINT: "{{ .Values.trace.jaeger.endpoint }}"
  TRACE_JAEGER_USERNAME: "{{ .Values.trace.jaeger.username }}"
  TRACE_JAEGER_AGENT_HOSTNAME: "{{ .Values.trace.jaeger.agent_host }}"
  TRACE_JAEGER_AGENT_PORT: "{{ .Values.trace.jaeger.agent_port }}"
  {{- else }}
  TRACE_OTEL_ENDPOINT: "{{ .Values.trace.otel.endpoint }}"
  TRACE_OTEL_URL_PATH: "{{ .Values.trace.otel.url_path }}"
  TRACE_OTEL_COMPRESSION: "{{ .Values.trace.otel.compression }}"
  TRACE_OTEL_INSECURE: "{{ .Values.trace.otel.insecure }}"
  TRACE_OTEL_TIMEOUT: "{{ .Values.trace.otel.timeout }}"
  {{- end }}
{{- end -}}

{{- define "harbor.traceEnvsForCore" -}}
  {{- if .Values.trace.enabled }}
  TRACE_SERVICE_NAME: "harbor-core"
  {{ include "harbor.traceEnvs" . }}
  {{- end }}
{{- end -}}

{{- define "harbor.traceEnvsForJobservice" -}}
  {{- if .Values.trace.enabled }}
  TRACE_SERVICE_NAME: "harbor-jobservice"
  {{ include "harbor.traceEnvs" . }}
  {{- end }}
{{- end -}}

{{- define "harbor.traceEnvsForRegistryCtl" -}}
  {{- if .Values.trace.enabled }}
  TRACE_SERVICE_NAME: "harbor-registryctl"
  {{ include "harbor.traceEnvs" . }}
  {{- end }}
{{- end -}}

{{- define "harbor.traceJaegerPassword" -}}
  {{- if and .Values.trace.enabled (eq .Values.trace.provider "jaeger") }}
  TRACE_JAEGER_PASSWORD: "{{ .Values.trace.jaeger.password | default "" | b64enc }}"
  {{- end }}
{{- end -}}

{{/* Allow KubeVersion to be overridden. */}}
{{- define "harbor.ingress.kubeVersion" -}}
  {{- default .Capabilities.KubeVersion.Version .Values.expose.ingress.kubeVersionOverride -}}
{{- end -}}
