package pkg

import (
	"net"
	"time"

	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
)

type VIPManager struct {
	id                  string
	bind                string
	fsm                 FSM
	peers               Peers
	logger              Logger
	stop                chan bool
	finished            chan bool
	networkConfigurator NetworkConfigurator
}

func NewVIPManager(id, bind string, peers Peers, logger Logger, networkConfigurator NetworkConfigurator) *VIPManager {
	return &VIPManager{id: id, peers: peers, bind: bind, fsm: FSM{}, logger: logger, networkConfigurator: networkConfigurator}
}

func (manager *VIPManager) addIP(verbose bool) {
	if error := manager.networkConfigurator.AddIP(); error != nil {
		log.WithFields(log.Fields{"error": error, "ip": manager.networkConfigurator.IP(), "interface": manager.networkConfigurator.Interface()}).Error("Could not set ip")
	} else if verbose {
		log.WithFields(log.Fields{"ip": manager.networkConfigurator.IP(), "interface": manager.networkConfigurator.Interface()}).Info("Added IP")
	}
}

func (manager *VIPManager) deleteIP(verbose bool) {
	if error := manager.networkConfigurator.DeleteIP(); error != nil {
		log.WithFields(log.Fields{"error": error, "ip": manager.networkConfigurator.IP(), "interface": manager.networkConfigurator.Interface()}).Error("Could not delete ip")
	} else if verbose {
		log.WithFields(log.Fields{"ip": manager.networkConfigurator.IP(), "interface": manager.networkConfigurator.Interface()}).Info("Deleted IP")
	}
}

func (manager *VIPManager) Start() error {
	// Create configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(manager.id)
	config.LogOutput = manager.logger

	// Initialize communication
	address, error := net.ResolveTCPAddr("tcp", manager.bind)
	if error != nil {
		return error
	}

	// Create transport
	transport, error := raft.NewTCPTransport(manager.bind, address, 3, 10*time.Second, manager.logger)
	if error != nil {
		return error
	}

	// Create Raft structures
	snapshots := raft.NewInmemSnapshotStore()
	logStore := raft.NewInmemStore()
	stableStore := raft.NewInmemStore()

	// Cluster configuration
	configuration := raft.Configuration{}

	for id, ip := range manager.peers {
		configuration.Servers = append(configuration.Servers, raft.Server{ID: raft.ServerID(id), Address: raft.ServerAddress(ip)})
	}

	// Bootstrap cluster
	if error := raft.BootstrapCluster(config, logStore, stableStore, snapshots, transport, configuration); error != nil {
		return error
	}

	// Create RAFT instance
	raftServer, error := raft.NewRaft(config, manager.fsm, logStore, stableStore, snapshots, transport)
	if error != nil {
		return error
	}

	manager.stop = make(chan bool, 1)
	manager.finished = make(chan bool, 1)
	ticker := time.NewTicker(time.Second)
	isLeader := false

	manager.deleteIP(false)

	go func() {
		for {
			select {
			case leader := <-raftServer.LeaderCh():
				if leader {
					isLeader = true

					log.Info("Leading")

					manager.addIP(true)

				} else {
					isLeader = false

					log.Info("Following")

					manager.deleteIP(true)
				}

			case <-ticker.C:
				if isLeader {
					result, error := manager.networkConfigurator.IsSet()
					if error != nil {
						log.WithFields(log.Fields{"error": error, "ip": manager.networkConfigurator.IP(), "interface": manager.networkConfigurator.Interface()}).Error("Could not check ip")
					}

					if result == false {
						log.Error("Lost IP")

						manager.addIP(true)
					}
				}

			case <-manager.stop:
				log.Info("Stopping")

				if isLeader {
					manager.deleteIP(true)
				}

				close(manager.finished)

				return
			}
		}
	}()

	log.Info("Started")

	return nil
}

func (manager *VIPManager) Stop() {
	close(manager.stop)

	<-manager.finished

	log.Info("Stopped")
}
