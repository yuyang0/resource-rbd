package rbd

import (
	"context"
	"testing"

	"github.com/projecteru2/core/log"
	"github.com/projecteru2/core/store/etcdv3/meta"
	coretypes "github.com/projecteru2/core/types"
)

const (
	name                = "rbd"
	rate                = 8
	nodeResourceInfoKey = "/resource/rbd/%s"
	priority            = -10000
)

// Plugin
type Plugin struct {
	name   string
	config coretypes.Config
	store  meta.KV
}

// NewPlugin .
func NewPlugin(ctx context.Context, cfg coretypes.Config, t *testing.T) (*Plugin, error) {
	if t == nil && len(cfg.Etcd.Machines) < 1 {
		return nil, coretypes.ErrConfigInvaild
	}
	var err error
	plugin := &Plugin{name: name, config: cfg}
	if plugin.store, err = meta.NewETCD(cfg.Etcd, t); err != nil {
		log.WithFunc("resource.rbd.NewPlugin").Error(ctx, err)
		return nil, err
	}
	return plugin, nil
}

// Name .
func (p Plugin) Name() string {
	return p.name
}
