---
apiVersion: v1
kind: Secret
metadata:
  name: wireguard
type: Opaque
stringData:
  wireguard: |
    [Interface]
    PrivateKey = {{ .Values.privateKey }}
    Address = {{ .Values.network }}
    ListenPort = {{ .Values.listenPort }}
    PostUp = iptables -A FORWARD -i {{ .Values.networkInterface }} -j ACCEPT ; iptables -t nat -A POSTROUTING -o {{ .Values.networkInterface }} -j MASQUERADE
    PostDown = iptables -D FORWARD -i {{ .Values.networkInterface }} -j ACCEPT; iptables -t nat -D POSTROUTING -o {{ .Values.networkInterface }} -j MASQUERADE

    {{- range $peerName, $peer := .Values.peers }}
    # Peer: {{ $peerName }}
    [Peer]
    PublicKey = {{ default $peer.pubkey $peer.publicKey }}
    AllowedIPs = {{ default $peer.allowedIPs $peer.src }}
    {{- end }}
  coredns: |
    . {
      forward . {{ .Values.dnsForward }}
      log
      errors
      cache
    }
