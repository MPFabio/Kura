{{/* Étiquettes standard appliquées à toutes les ressources. */}}
{{- define "kura.labels" -}}
app.kubernetes.io/part-of: kura
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version }}
{{- end }}

{{/* Étiquettes de sélection pour un composant donné (passer le nom en argument). */}}
{{- define "kura.selectorLabels" -}}
app.kubernetes.io/name: {{ . }}
app.kubernetes.io/instance: kura
{{- end }}

{{/* Référence d'image d'un microservice Kura dans le registre du client. */}}
{{- define "kura.serviceImage" -}}
{{ printf "%s/kura/%s:%s" .root.Values.global.imageRegistry .name .root.Values.global.imageTag }}
{{- end }}

