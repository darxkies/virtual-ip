package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"

	"github.com/darxkies/virtual-ip/pkg"
	"github.com/darxkies/virtual-ip/version"
	log "github.com/sirupsen/logrus"
)

func main() {

	id := flag.String("id", "vip", "ID of this node")
	bind := flag.String("bind", "0.0.0.0", "RAFT bind addreess")
	peersList := flag.String("peers", "", "Peers as a comma separated list of peer-id=peer-address:peer-port including the id and the bind of this instance")
	_interface := flag.String("interface", "lo", "Network interface")
	virtualIP := flag.String("virtual-ip", "192.168.0.25", "Virtual/Floating IP")
	flag.Parse()

	log.WithFields(log.Fields{"version": version.Version}).Info("Virtual-IP")
	log.WithFields(log.Fields{"id": *id, "bind": *bind, "peers": *peersList}).Info("Cluster")
	log.WithFields(log.Fields{"virtual-ip": *virtualIP, "interface": *_interface}).Info("Network")

	netlinkNetworkConfigurator, error := pkg.NewNetlinkNetworkConfigurator(*virtualIP, *_interface)
	if error != nil {
		log.WithFields(log.Fields{"error": error}).Error("Network failure")

		os.Exit(-1)
	}

	peers := pkg.Peers{}

	if len(*peersList) > 0 {
		for _, peer := range strings.Split(*peersList, ",") {
			peerTokens := strings.Split(peer, "=")

			if len(peerTokens) != 2 {
				log.WithFields(log.Fields{"peer": peer}).Error("Peers malformated")

				os.Exit(-1)
			}

			peers[peerTokens[0]] = peerTokens[1]
		}
	}

	logger := pkg.Logger{}

	vipManager := pkg.NewVIPManager(*id, *bind, peers, logger, netlinkNetworkConfigurator)
	if error := vipManager.Start(); error != nil {
		log.WithFields(log.Fields{"error": error}).Error("Start failed")

		os.Exit(-1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan

	vipManager.Stop()
}
