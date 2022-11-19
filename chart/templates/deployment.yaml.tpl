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
          imagePullPolicy: Always
          env:
            - name: WG_INTERFACE
              value: "{{ .Values.interface }}"
            - name: WG_IP_FORWARDING
              value: "true"
            - name: WG_ENABLE
              value: "true"
            - name: WG_COREDNS
              value: "true"
            {{- if .Values.metrics }}
            - name: WG_PEER_MONITOR
              value: "true"
            - name: WG_PROM_METRICS
              value: "true"
            {{- end }}
          securityContext:
            privileged: true
            capabilities:
              add:
                - NET_ADMIN
                - NET_BIND_SERVICE
          ports:
            - containerPort: {{ .Values.listenPort }}
              protocol: UDP
            - containerPort: 8080
              protocol: TCP
            - containerPort: 9090
              protocol: TCP
            - containerPort: 53
              protocol: UDP
          resources:
            limits:
              memory: 512Mi
            requests:
              cpu: 0.25
              memory: 128Mi
          volumeMounts:
            - name: wireguard-config
              mountPath: "/etc/wireguard/{{ .Values.interface }}.conf"
              subPath: wireguard
              readOnly: true
            - name: wireguard-config
              mountPath: /etc/coredns/Corefile
              subPath: coredns
              readOnly: true
            - name: wireguard-config
              mountPath: /etc/coredns/private_hosts
              subPath: coredns_private_hosts
              readOnly: true
      volumes:
        - name: wireguard-config
          secret:
            secretName: wireguard
