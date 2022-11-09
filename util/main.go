package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
	iface := os.Getenv("WG_INTERFACE")
	if iface == "" {
		iface = "wg0"
	}

	if os.Getenv("WG_IP_FORWARDING") != "" {
		log.Println("enabling ip forwarding via sysctl")
		if err := enableIPForwarding(); err != nil {
			log.Fatal("sysctl command failed:", err)
		}
	}

	if os.Getenv("WG_ENABLE") != "" {
		log.Println("enabling wireguard interface", iface)
		if err := runCmd("wg-quick", "up", iface); err != nil {
			log.Fatal("failed to bring interface up:", err)
		}
	}

	http.HandleFunc("/health", func(rw http.ResponseWriter, req *http.Request) {
		out, err := runWithOutput("wg", "show", iface)
		if err != nil {
			log.Println("wireguard healthcheck failed:", out)
			rw.WriteHeader(500)
			fmt.Fprintf(rw, "UNHEALTHY")
			return
		}

		rw.WriteHeader(200)
		fmt.Fprintf(rw, "HEALTHY")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("starting web endpoint on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func enableIPForwarding() error {
	if err := runCmd("sysctl", "-w", "net.ipv4.ip_forward=1"); err != nil {
		return err
	}
	return runCmd("sysctl", "-w", "net.ipv4.conf.all.forwarding=1")
}
