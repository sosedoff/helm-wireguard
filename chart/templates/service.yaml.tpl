---
apiVersion: v1
kind: Service
metadata:
  name: wireguard
spec:
  selector:
    app: wireguard
  type: ClusterIP
  ports:
    - name: http
      protocol: TCP
      port: 8080
      targetPort: 8080
    - name: wireguard
      protocol: UDP
      port: {{ .Values.listenPort }}
      targetPort: {{ .Values.listenPort }}
    - name: coredns
      protocol: UDP
      port: 53
      targetPort: 53
