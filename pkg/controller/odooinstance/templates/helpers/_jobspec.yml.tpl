
{{- define "jobspec" }}
spec:
  completions: 1
  backoffLimit: 1
  activeDeadlineSeconds: 360
  template:
    metadata:
      labels:
        cluster.odoo.io/name: {{ .Instance.Spec.Cluster | quote }}
        cluster.odoo.io/track: {{ .Extra.Track | quote }}
        instance.odoo.io/hostname: {{ .Instance.Spec.Hostname | quote }}
        app.kubernetes.io/name: {{ block "componentName" . }}{{ end }}
        app.kubernetes.io/instance: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
        app.kubernetes.io/component: {{ block "componentType" . }}{{ end }}
        app.kubernetes.io/managed-by: odoo-operator
        app.kubernetes.io/part-of: {{ .Instance.Name | quote }}
        app.kubernetes.io/version: {{ .Instance.Spec.Version | quote }}
    spec:
      restartPolicy: Never
      imagePullSecrets:
      - name: pull-secret
      terminationMessagePolicy: FallbackToLogsOnError
      securityContext:
        fsGroup: 9001
        runAsUser: 9001
        runAsGroup: 9001
        runAsNonRoot: true
        supplementalGroups: [2000]
      containers:
      - name: default
        image: {{ .Extra.Image }}:devops-{{ .Instance.Spec.Version }}
        imagePullPolicy: Always
        args:
        {{ block "jobArgs" . }}{{ end }}
        resources:
          requests:
            memory: 512M
            cpu: 200m
          limits:
            memory: 1G
            cpu: 500m
        env:
         - name: PGHOST
           value: {{ .Extra.Database.Host }}
         - name: PGUSER
           value: {{ .Extra.Database.User }}
         - name: PGPORT
           value: {{ .Extra.Database.Port }}
         - name: PGPASSWORD
           value: {{ .Extra.Database.Password }}
         - name: ODOO_RC
           value: /run/configs/odoo/
         - name: ODOO_PASSFILE
           value: /run/secrets/odoo/adminpwd
        volumeMounts:
        - name: data-volume
          mountPath: /mnt/odoo/data/
        - name: config-volume
          mountPath: /run/configs/odoo/
          readonly: true
        - name: app-secrets
          mountPath: /run/secrets/odoo/
          readonly: true

      volumes:
        - name: data-volume
          configMap:
            name: {{ .Extra.ClusterName }}-data-volume
        - name: config-volume
          configMap:
            name: {{ .Instance.Name }}-config
        - name: app-secrets
          secret:
            secretName: {{ .Extra.ClusterName }}-{{ .Instance.Spec.Version }}-secret
{{ end -}}