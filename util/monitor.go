package main

import (
	"context"
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
	inactiveTxTreshold        = 1 * time.Minute
)

var (
	peerMap   = map[string]*wgtypes.Peer{}
	peerTxMap = map[string]time.Time{}
)

type peerStats struct {
	count  int
	active int
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
		case <-ticker.C:
			stats, err := checkPeers(client, iface)
			if err != nil {
				log.Println("monitor error:", err)
				break
			}
			log.Printf("peer monitor: total=%v active=%v\n", stats.count, stats.active)
		case <-ctx.Done():
			log.Println("stopping peer monitor")
			return
		}
	}
}

func checkPeers(client *wgctrl.Client, iface string) (stats peerStats, err error) {
	device, err := client.Device(iface)
	if err != nil {
		return stats, err
	}

	stats.count = len(device.Peers)

	for _, peer := range device.Peers {
		key := peer.PublicKey.String()

		lastPeer := peerMap[key]
		if lastPeer != nil && isPeerActive(lastPeer, &peer) {
			stats.active++
		}

		peerMap[key] = &peer
	}

	return stats, nil
}

func isPeerActive(prevPeer, peer *wgtypes.Peer) bool {
	// Peer is known but has not made any connections yet
	if peer.LastHandshakeTime.IsZero() {
		return false
	}

	// Newly seen peer that might not have correct handshake time
	prevHandshakeTime := prevPeer.LastHandshakeTime
	if prevHandshakeTime.IsZero() {
		prevHandshakeTime = peer.LastHandshakeTime
	}

	key := peer.PublicKey.String()
	now := time.Now()

	// Peer is actively sending or receiving data
	if peer.ReceiveBytes > prevPeer.ReceiveBytes ||
		peer.TransmitBytes > prevPeer.TransmitBytes {
		peerTxMap[key] = now
		return true
	}

	handshakeDiff := peer.LastHandshakeTime.Sub(prevHandshakeTime)
	txDiff := now.Sub(peerTxMap[key])

	// Peer handshake has expired but still sending data
	if handshakeDiff > inactiveHandshakeTreshold {
		return txDiff > inactiveTxTreshold
	}

	// Detemine active status based on last data transfer
	return txDiff < inactiveTxTreshold
}
