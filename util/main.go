package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.zx2c4.com/wireguard/wgctrl"
)

func main() {
	cmd := "start"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	switch cmd {
	case "generate":
		generateServerConfig()
	case "generate-peer":
		generatePeerConfig()
	case "start":
		start()
	default:
		log.Fatal("unknown command:", cmd)
	}
}

func start() {
	iface := getEnvVarWithDefault("WG_INTERFACE", "wg0")

	if isEnvVarSet("WG_ENABLE") {
		log.Println("enabling wireguard interface", iface)
		if err := runCmd("wg-quick", "up", iface); err != nil {
			log.Fatal("failed to bring interface up:", err)
		}
	}

	if isEnvVarSet("WG_PEER_MONITOR") {
		monitor, err := NewMonitor(iface)
		if err != nil {
			log.Fatal("cant start monitor:", err)
		}
		go monitor.Start(context.Background())
	}

	if isEnvVarSet("WG_PROM_METRICS") {
		promAddr := getEnvVarWithDefault("WG_PROM_ADDR", ":9090")
		go startMetricsServer(context.Background(), promAddr)
	}

	if isEnvVarSet("WG_COREDNS") {
		log.Println("starting coredns")
		go startDns(context.Background())
	}

	startServer(iface)
}

func startServer(iface string) {
	client, err := wgctrl.New()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/health", func(rw http.ResponseWriter, req *http.Request) {
		_, err := client.Device(iface)
		if err != nil {
			log.Println("cant obtain device:", err)
			rw.WriteHeader(500)
			fmt.Fprintf(rw, "UNHEALTHY")
			return
		}

		rw.WriteHeader(200)
		fmt.Fprintf(rw, "HEALTHY")
	})

	port := getEnvVarWithDefault("PORT", "8080")
	log.Println("starting web endpoint on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func isEnvVarSet(key string) bool {
	val := strings.ToLower(os.Getenv(key))
	return val == "1" || val == "true"
}

func getEnvVarWithDefault(key string, defval string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defval
	}
	return val
}

func startDns(ctx context.Context) error {
	return runCmd("coredns", "-conf", "/etc/coredns/Corefile")
}
