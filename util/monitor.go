package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	// How often to run monitor
	monitorSchedule = 5 * time.Second

	// Peer to be considered inactive based on last handshake or data transmission time
	inactiveHandshakeTreshold = 3 * time.Minute
	inactiveTxTreshold        = 30 * time.Second
)

var (
	peerMap      = map[string]*wgtypes.Peer{}
	peerStateMap = map[string]peerInfo{}
)

type peerInfo struct {
	lastTxTime time.Time
	lastRxTime time.Time
}

type peerStats struct {
	iface      string
	count      int
	active     int
	txSent     int64
	txReceived int64
}

func (stats peerStats) String() string {
	return fmt.Sprintf("peer stats: total=%v active=%v tx_rate=%v rx_rate=%v",
		stats.count, stats.active,
		stats.txReceived, stats.txSent,
	)
}

func startPeersMonitor(ctx context.Context, iface string) {
	log.Println("starting peers monitor for", iface)

	client, err := wgctrl.New()
	if err != nil {
		log.Fatal("monitor failed to obtain wireguard client:", err)
	}

	ticker := time.NewTicker(monitorSchedule)
	defer ticker.Stop()

	for {
		select {
		case ts := <-ticker.C:
			stats, err := checkPeers(client, iface, ts)
			if err != nil {
				log.Println("monitor error:", err)
				break
			}

			setMetricsFromStats(stats)
			setPeerMetrics(iface, ts)

			log.Println(stats)
		case <-ctx.Done():
			log.Println("stopping peer monitor")
			return
		}
	}
}

func setMetricsFromStats(stats peerStats) {
	activePeerGauge.WithLabelValues(stats.iface).Set(float64(stats.active))
	totalPeerGauge.WithLabelValues(stats.iface).Set(float64(stats.count))
	txReceivedCounter.WithLabelValues(stats.iface).Add(float64(stats.txReceived))
	txSentCounter.WithLabelValues(stats.iface).Add(float64(stats.txSent))
}

func setPeerMetrics(iface string, ts time.Time) {
	for peerKey, peer := range peerStateMap {
		if peerInfo := peerMap[peerKey]; peerInfo != nil {
			peerLastHandshakeGauge.WithLabelValues(iface, peerKey).Set(ts.Sub(peerInfo.LastHandshakeTime).Seconds())
			peerLastTxGauge.WithLabelValues(iface, peerKey).Set(ts.Sub(peer.lastTxTime).Seconds())
			peerLastRxGauge.WithLabelValues(iface, peerKey).Set(ts.Sub(peer.lastRxTime).Seconds())
		}
	}
}

func checkPeers(client *wgctrl.Client, iface string, ts time.Time) (stats peerStats, err error) {
	device, err := client.Device(iface)
	if err != nil {
		return stats, err
	}

	stats.iface = iface
	stats.count = len(device.Peers)

	for _, peer := range device.Peers {
		key := peer.PublicKey.String()

		lastPeer := peerMap[key]
		if lastPeer != nil {
			if isPeerActive(lastPeer, &peer, ts) {
				stats.active++
			}

			stats.txReceived = stats.txReceived + peer.ReceiveBytes - lastPeer.ReceiveBytes
			stats.txSent = stats.txSent + peer.TransmitBytes - lastPeer.TransmitBytes
		}

		peerMap[key] = &peer
	}

	return stats, nil
}

func isPeerActive(prevPeer, peer *wgtypes.Peer, ts time.Time) bool {
	// Peer is known but has not made any connections yet
	if peer.LastHandshakeTime.IsZero() {
		return false
	}

	key := peer.PublicKey.String()
	checkTxRxTimes := true

	// Record peer tx/rx timestamps
	info := peerStateMap[key]
	if info.lastRxTime.IsZero() && info.lastTxTime.IsZero() {
		info.lastRxTime = ts
		info.lastTxTime = ts
		peerStateMap[key] = info
		checkTxRxTimes = false
	}

	// Save last observed TX/RX timestamps
	if peer.ReceiveBytes > prevPeer.ReceiveBytes {
		info.lastRxTime = ts
	}
	if peer.TransmitBytes > prevPeer.TransmitBytes {
		info.lastTxTime = ts
	}
	peerStateMap[key] = info

	if checkTxRxTimes {
		// Mark peer as active since we observed changes in both TX/RX timestamps
		if ts.Sub(info.lastRxTime) <= inactiveTxTreshold && ts.Sub(info.lastTxTime) <= inactiveTxTreshold {
			return true
		}
	}

	// ---------------------------------------------------------------------------
	// NOTE: cannot determine active state based on traffic patterns, due to the fact
	// that while peer is disconnected (via the UI or CLI) it might still send
	// "some" packets to wireguard. In that scenario the only way to know for sure
	// if the peer is active based on the last handshake, which occurs ever ~2mins.
	// Only tested on Wireguard for Mac (which is based on wireguard-go).
	// ---------------------------------------------------------------------------

	// Return activity status based on the last handshake timestamp
	return ts.Sub(peer.LastHandshakeTime) < inactiveHandshakeTreshold
}
