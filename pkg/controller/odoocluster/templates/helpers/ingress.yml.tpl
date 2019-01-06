{{ define "ingress" }}
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
  namespace: {{ .Instance.Namespace }}
  labels:
    app.kubernetes.io/name: {{ block "componentName" . }}{{ end }}
    app.kubernetes.io/instance: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
    app.kubernetes.io/version: {{ .Instance.Spec.Version }}
    app.kubernetes.io/component: {{ block "componentType" . }}{{ end }}
    app.kubernetes.io/part-of: {{ .Instance.Name }}
    app.kubernetes.io/managed-by: odoo-operator
  annotations:
    kubernetes.io/ingress.class: traefik
    kubernetes.io/tls-acme: "true"  # Still support kube-lego in addition to certmanager
    certmanager.k8s.io/cluster-issuer: letsencrypt-prod
spec:
  rules:
  - host: {{ .Instance.Spec.Hostname }}
    http:
      paths:
      - path: {{ block "ingressPath" . }}{{ end }}
        backend:
          serviceName: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
          servicePort: 8000
  tls:
  - secretName: {{ .Instance.Spec.Secret }}-tls
    hosts:
    - {{ .Instance.Spec.Hostname }}
{{ end }}