{{- define "name" -}}
gardener-extension-csi-driver-synology
{{- end -}}

{{- define "labels" -}}
app.kubernetes.io/name: {{ include "name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
