package main

import (
	"bytes"
	"fmt"
	"log"
	"text/template"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const valuesTemplate = `
# ------------------------------------------------------------------------------
# This is your wireguard helm chart config
# ------------------------------------------------------------------------------
image: sosedoff/wireguard:latest
network: 10.10.0.0/24
privateKey: {{ .serverKey }}
publicKey: {{ .serverPubkey }}
dns: true
`

const clientTemplate = `
[Interface]
PrivateKey = {{ .key }}
Address = 10.10.0.1/32
DNS = 10.10.0.0

[Peer]
PublicKey = <SERVER_PUBKEY>
AllowedIPs = 0.0.0.0/0
Endpoint = <SERVER_IP>:51820
PersistentKeepalive = 25
`

const peerTemplate = `
username:
  privateKey: {{ .key }}
  publicKey: {{ .pubKey }}
  src: 10.10.X.X/32
`

func generateServerConfig() {
	tpl, err := template.New("config").Parse(valuesTemplate)
	if err != nil {
		log.Fatal("cant parse template:", err)
	}

	serverKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Fatal("unable to generate server private key:", err)
	}

	out := bytes.NewBuffer(nil)
	err = tpl.Execute(out, map[string]interface{}{
		"serverKey":    serverKey.String(),
		"serverPubkey": serverKey.PublicKey().String(),
	})

	fmt.Printf("%s\n", out.String())
}

func generatePeerConfig() {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Fatal("unable to generate client key:", err)
	}

	tpl, err := template.New("peer").Parse(clientTemplate)
	if err != nil {
		log.Fatal("cant parse template:", err)
	}

	peerTpl, err := template.New("values").Parse(peerTemplate)
	if err != nil {
		log.Fatal("cant parse template:", err)
	}

	out := bytes.NewBuffer(nil)
	err = peerTpl.Execute(out, map[string]interface{}{
		"key":    key.String(),
		"pubKey": key.PublicKey().String(),
	})
	fmt.Println("> New peer for helm chart (replace values where needed):")
	fmt.Printf("%s\n", out.String())

	out.Reset()
	err = tpl.Execute(out, map[string]interface{}{
		"key": key.String(),
	})
	fmt.Println("> Wireguard client config (replace values where needed):")
	fmt.Printf("%s\n", out.String())
	fmt.Println("# ==============================================================")
}
