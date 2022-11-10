---
apiVersion: apps/v1
kind: {{ .Values.deployKind }}
metadata:
  name: wireguard
  labels:
    app: wireguard
spec:
  {{- if eq .Values.deployKind "Deployment" }}
  replicas: {{ default 1 .Values.replicas }}
  {{- end }}
  selector:
    matchLabels:
      app: wireguard
  template:
    metadata:
      labels:
        app: wireguard
    spec:
      restartPolicy: Always
      terminationGracePeriodSeconds: 30
      automountServiceAccountToken: false
      containers:
        - name: wireguard
          image: {{ .Values.image }}
          env:
            - name: WG_INTERFACE
              value: "{{ .Values.interface }}"
            - name: WG_IP_FORWARDING
              value: "true"
            - name: WG_ENABLE
              value: "true"
            {{- if .Values.metrics }}
            - name: WG_PEER_MONITOR
              value: "true"
            {{- end }}
          securityContext:
            privileged: true
            capabilities:
              add:
                - NET_ADMIN
          ports:
            - containerPort: {{ .Values.listenPort }}
              protocol: UDP
          volumeMounts:
            - name: wireguard-config
              mountPath: "/etc/wireguard/{{ .Values.interface }}.conf"
              subPath: wireguard
              readOnly: true
        - name: coredns
          image: {{ .Values.dnsImage }}
          args:
            - "-conf"
            - /etc/coredns/Corefile
          ports:
            - containerPort: 53
              protocol: UDP
          securityContext:
            privileged: true
            capabilities:
              add:
                - NET_BIND_SERVICE
          volumeMounts:
            - name: wireguard-config
              mountPath: /etc/coredns/Corefile
              subPath: coredns
              readOnly: true
      volumes:
        - name: wireguard-config
          secret:
            secretName: wireguard
