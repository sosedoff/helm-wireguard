package main

import (
	"context"
	"log"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	// Peer will be considered inactive if last handshake was received over X minutes
	inactiveHandshakeTreshold = 3 * time.Minute
)

func startPeersMonitor(ctx context.Context, iface string) {
	log.Println("starting peers monitor for", iface)

	client, err := wgctrl.New()
	if err != nil {
		log.Fatal("monitor failed to obtain wireguard client:", err)
	}

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	peerMap := map[string]*wgtypes.Peer{}

	checkPeers := func() {
		device, err := client.Device(iface)
		if err != nil {
			log.Println("monitor failed to get device for", iface)
		}

		numTotal := len(device.Peers)
		numActive := 0

		for _, peer := range device.Peers {
			key := peer.PublicKey.String()

			lastPeer := peerMap[key]
			if lastPeer != nil && isPeerActive(lastPeer, &peer) {
				numActive++
			}

			peerMap[key] = &peer
		}

		log.Printf("peers total=%v active=%v\n", numTotal, numActive)
	}

	for {
		select {
		case <-ticker.C:
			checkPeers()
		case <-ctx.Done():
			log.Println("stopping peer monitor")
			return
		}
	}
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

	// Mark as inactive when has not seen for over N minutes
	diff := peer.LastHandshakeTime.Sub(prevHandshakeTime)
	if diff > inactiveHandshakeTreshold {
		return false
	}

	return peer.ReceiveBytes >= prevPeer.ReceiveBytes ||
		peer.TransmitBytes >= prevPeer.TransmitBytes
}
