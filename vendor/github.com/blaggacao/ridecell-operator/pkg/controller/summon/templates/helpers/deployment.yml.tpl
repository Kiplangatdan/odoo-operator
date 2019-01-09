{{ define "deployment" }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
  namespace: {{ .Instance.Namespace }}
  labels:
    app.kubernetes.io/name: {{ block "componentName" . }}{{ end }}
    app.kubernetes.io/instance: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
    app.kubernetes.io/version: {{ .Instance.Spec.Version }}
    app.kubernetes.io/component: {{ block "componentType" . }}{{ end }}
    app.kubernetes.io/part-of: {{ .Instance.Name }}
    app.kubernetes.io/managed-by: summon-operator
spec:
  replicas: {{ block "replicas" . }}1{{ end }}
  selector:
    matchLabels:
      app.kubernetes.io/instance: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ block "componentName" . }}{{ end }}
        app.kubernetes.io/instance: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
        app.kubernetes.io/version: {{ .Instance.Spec.Version }}
        app.kubernetes.io/component: {{ block "componentType" . }}{{ end }}
        app.kubernetes.io/part-of: {{ .Instance.Name }}
        app.kubernetes.io/managed-by: summon-operator
    spec:
      imagePullSecrets:
      - name: pull-secret
      containers:
      - name: default
        image: us.gcr.io/ridecell-1/summon:{{ .Instance.Spec.Version }}
        imagePullPolicy: Always
        command: {{ block "command" . }}[]{{ end }}
        ports: {{ block "deploymentPorts" . }}[{containerPort: 8000}]{{ end }}
        resources:
          requests:
            memory: 512M
            cpu: 500m
          limits:
            memory: 1G
            cpu: 1000m
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
        - name: app-secrets
          mountPath: /etc/secrets

      volumes:
        - name: config-volume
          configMap:
            name: {{ .Instance.Name }}-config
        - name: secrets-orig
          secret:
            secretName: {{ .Instance.Spec.Secret }}
        - name: app-secrets
          secret:
            secretName: summon.{{ .Instance.Name }}.app-secrets
{{ end }}
