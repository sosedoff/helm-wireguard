package main

import (
	"context"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	baseLabels = []string{"interface"}
	peerLabels = []string{"interface", "peer"}

	totalPeerGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peers",
		Namespace: "wireguard",
		Help:      "Number of peers in configuration",
	}, baseLabels)

	activePeerGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "active_peers",
		Namespace: "wireguard",
		Help:      "Number of active peers",
	}, baseLabels)

	txReceivedGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "tx_received",
		Namespace: "wireguard",
		Help:      "Total number of bytes received on the interface",
	}, baseLabels)

	txSentGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "tx_sent",
		Namespace: "wireguard",
		Help:      "Total number of bytes sent on the interface",
	}, baseLabels)

	peerLastHandshakeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peer_last_handshake_seconds",
		Namespace: "wireguard",
		Help:      "Number of seconds since last handshake",
	}, peerLabels)

	peerLastTxGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peer_last_tx_seconds",
		Namespace: "wireguard",
		Help:      "Number of seconds since last TX activity",
	}, peerLabels)

	peerLastRxGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peer_last_rx_seconds",
		Namespace: "wireguard",
		Help:      "Number of seconds since last RX activity",
	}, peerLabels)
)

func startMetricsServer(ctx context.Context, addr string) error {
	log.Println("starting prometheus metrics at:", addr)

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		activePeerGauge, totalPeerGauge,
		txReceivedGauge, txSentGauge,
		peerLastHandshakeGauge, peerLastTxGauge, peerLastRxGauge,
	)

	handler := promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: false,
		},
	)

	http.Handle("/metrics", handler)
	return http.ListenAndServe(addr, nil)
}
