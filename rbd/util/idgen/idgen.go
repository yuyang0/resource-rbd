package idgen

import (
	"github.com/bwmarrin/snowflake"
)

var (
	node *snowflake.Node
)

func Init(nodeid int64) error {
	var err error
	node, err = snowflake.NewNode(nodeid)
	return err
}

func GenID() string {
	id := node.Generate()
	return id.Base58()
}
