package testmgrsrv

import (
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
)

type TestConfig struct {
	db    dbagent.DBAgent
	Nodes *map[string]dtype.NodeInfo
}

// Select transaction Randomly
func (tc *TestConfig) GetANwithRandom() *dtype.NodeInfo {
	return nil
}

// Select latest transactions more
func (tc *TestConfig) GetANwithTimeWeight() *dtype.NodeInfo {
	return nil
}

func NewTestConfig(db dbagent.DBAgent, nodes *map[string]dtype.NodeInfo) *TestConfig {
	tc := &TestConfig{
		db:    db,
		Nodes: nodes,
	}

	return tc
}
