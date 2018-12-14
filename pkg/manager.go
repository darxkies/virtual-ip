package pkg

import (
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/raft"
	log "github.com/sirupsen/logrus"
)

type VIPManager struct {
	id            string
	bind          string
	virtualIP     string
	fsm           FSM
	peers         Peers
	logger        Logger
	_interface    string
	stop          chan bool
	commandRunner CommandRunner
}

func NewVIPManager(id, bind string, virtualIP string, peers Peers, logger Logger, _interface string, commandRunner CommandRunner) *VIPManager {
	return &VIPManager{id: id, peers: peers, bind: bind, virtualIP: virtualIP, fsm: FSM{}, logger: logger, _interface: _interface, commandRunner: commandRunner}
}

func (manager *VIPManager) updateNetworkConfiguration(action string) error {
	command := fmt.Sprintf("ip addr %s %s/32 dev %s", action, manager.virtualIP, manager._interface)

	if error := manager.commandRunner.Run(command); error != nil {
		log.WithFields(log.Fields{"action": action, "error": error}).Error("Network update failed")

		return error
	}

	return nil
}

func (manager *VIPManager) addIP() error {
	log.Info("Add virtual ip")

	return manager.updateNetworkConfiguration("add")
}

func (manager *VIPManager) deleteIP() error {
	log.Info("Delete virtual ip")

	return manager.updateNetworkConfiguration("delete")
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

	_ = manager.deleteIP()

	go func() {
		for {
			select {
			case leader := <-raftServer.LeaderCh():
				if leader {
					_ = manager.addIP()
				} else {
					_ = manager.deleteIP()
				}

			case <-manager.stop:
				_ = manager.deleteIP()
			}
		}
	}()

	log.Info("Started")

	return nil
}

func (manager *VIPManager) Stop() {
	close(manager.stop)

	log.Info("Stopped")
}
