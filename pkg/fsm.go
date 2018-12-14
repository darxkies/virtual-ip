package pkg

import (
	"io"

	"github.com/hashicorp/raft"
)

type FSM struct {
}

func (fsm FSM) Apply(log *raft.Log) interface{} {
	return nil
}

func (fsm FSM) Restore(snap io.ReadCloser) error {
	return nil
}

func (fsm FSM) Snapshot() (raft.FSMSnapshot, error) {
	return Snapshot{}, nil
}
