package pkg

import "github.com/hashicorp/raft"

type Snapshot struct {
}

func (snapshot Snapshot) Persist(sink raft.SnapshotSink) error {
	return nil
}

func (snapshot Snapshot) Release() {
}
