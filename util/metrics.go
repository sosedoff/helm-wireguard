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
		Name:      "peers_count",
		Namespace: "wireguard",
		Help:      "Total number of peers in the interface configuration",
	}, baseLabels)

	activePeerGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peers_active_count",
		Namespace: "wireguard",
		Help:      "Total number of active peers",
	}, baseLabels)

	ifaceBytesTxGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "bytes_sent_total",
		Namespace: "wireguard",
		Help:      "Total number of bytes sent on the interface",
	}, baseLabels)

	ifaceBytesRxGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "bytes_received_total",
		Namespace: "wireguard",
		Help:      "Total number of bytes received on the interface",
	}, baseLabels)

	peerTxGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peer_bytes",
		Namespace: "wireguard",
		Help:      "Total number of bytes transmitted to this peer",
	}, peerLabels)

	peerRxGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peer_rx_bytes",
		Namespace: "wireguard",
		Help:      "Total number of bytes received from this peer",
	}, peerLabels)

	peerLastHandshakeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peer_last_handshake_seconds",
		Namespace: "wireguard",
		Help:      "Number of seconds since last handshake from this peer",
	}, peerLabels)

	peerLastTxGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peer_last_trasmit_seconds",
		Namespace: "wireguard",
		Help:      "Number of seconds since data trasmitted to this peer",
	}, peerLabels)

	peerLastRxGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "peer_last_receive_seconds",
		Namespace: "wireguard",
		Help:      "Number of seconds since data received from this peer",
	}, peerLabels)
)

func startMetricsServer(ctx context.Context, addr string) error {
	log.Println("starting prometheus metrics at:", addr)

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		activePeerGauge,
		totalPeerGauge,
		ifaceBytesTxGauge,
		ifaceBytesRxGauge,
		peerLastHandshakeGauge,
		peerLastTxGauge,
		peerLastRxGauge,
		peerTxGauge,
		peerRxGauge,
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
