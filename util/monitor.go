package main

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	// How often to run monitor
	monitorSchedule = 5 * time.Second

	// Peer to be considered inactive based on last handshake or data transmission time
	inactiveHandshakeTreshold = 180 * time.Second
)

type (
	PeerInfo struct {
		LastTxTime  time.Time
		LastTxBytes int64
		LastRxTime  time.Time
		LastRxBytes int64
	}

	Monitor struct {
		iface  string
		period time.Duration
		client *wgctrl.Client
		peerTx map[wgtypes.Key]*PeerInfo
	}
)

func NewMonitor(iface string) (*Monitor, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}

	return &Monitor{
		iface:  iface,
		client: client,
		peerTx: map[wgtypes.Key]*PeerInfo{},
		period: monitorSchedule,
	}, nil
}

func (m *Monitor) Start(ctx context.Context) error {
	log.Printf("[%s] starting monitor\n", m.iface)

	ticker := time.NewTicker(m.period)
	defer ticker.Stop()

	for {
		select {
		case ts := <-ticker.C:
			if err := m.run(ts); err != nil {
				log.Printf("[%s] monitor error: %v\n", m.iface, err)
			}
		case <-ctx.Done():
			log.Printf("[%s] stopping monitor", m.iface)
			return nil
		}
	}
}

func (m *Monitor) run(ts time.Time) error {
	device, err := m.client.Device(m.iface)
	if err != nil {
		return err
	}

	activePeers := 0
	totalPeers := len(device.Peers)
	bytesTx := int64(0)
	bytesRx := int64(0)

	for _, peer := range device.Peers {
		bytesTx = bytesTx + peer.TransmitBytes
		bytesRx = bytesRx + peer.ReceiveBytes

		if ts.Sub(peer.LastHandshakeTime) < inactiveHandshakeTreshold {
			activePeers++
		}

		txInfo := m.peerTx[peer.PublicKey]
		if txInfo != nil {
			if peer.TransmitBytes > txInfo.LastTxBytes {
				txInfo.LastTxBytes = peer.TransmitBytes
				txInfo.LastTxTime = ts
			}
			if peer.ReceiveBytes > txInfo.LastRxBytes {
				txInfo.LastRxBytes = peer.ReceiveBytes
				txInfo.LastRxTime = ts
			}
		} else {
			m.peerTx[peer.PublicKey] = &PeerInfo{}
		}

		m.setPeerMetrics(peer, ts)
	}

	activePeerGauge.WithLabelValues(m.iface).Set(float64(activePeers))
	totalPeerGauge.WithLabelValues(m.iface).Set(float64(totalPeers))
	ifaceBytesTxGauge.WithLabelValues(m.iface).Set(float64(bytesTx))
	ifaceBytesRxGauge.WithLabelValues(m.iface).Set(float64(bytesRx))

	return nil
}

func (m *Monitor) setPeerCounts(total, active int) {
	labels := prometheus.Labels{"interface": m.iface}

	totalPeerGauge.With(labels).Set(float64(total))
	activePeerGauge.With(labels).Set(float64(total))
}

func (m *Monitor) setPeerMetrics(peer wgtypes.Peer, ts time.Time) {
	labels := prometheus.Labels{"interface": m.iface, "peer": peer.PublicKey.String()}

	if !peer.LastHandshakeTime.IsZero() {
		peerLastHandshakeGauge.With(labels).Set(ts.Sub(peer.LastHandshakeTime).Seconds())
	}

	peerTxGauge.With(labels).Set(float64(peer.TransmitBytes))
	peerRxGauge.With(labels).Set(float64(peer.ReceiveBytes))

	if tx := m.peerTx[peer.PublicKey]; tx != nil {
		if !tx.LastRxTime.IsZero() {
			peerLastTxGauge.With(labels).Set(ts.Sub(tx.LastTxTime).Seconds())
		}
		if !tx.LastRxTime.IsZero() {
			peerLastRxGauge.With(labels).Set(ts.Sub(tx.LastRxTime).Seconds())
		}
	}
}
