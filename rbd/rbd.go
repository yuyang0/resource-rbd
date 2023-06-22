package rbd

import (
	"context"
	"testing"

	"github.com/projecteru2/core/log"
	"github.com/projecteru2/core/store/etcdv3/meta"
	coretypes "github.com/projecteru2/core/types"

	"github.com/yuyang0/resource-rbd/config"
	rbdapi "github.com/yuyang0/resource-rbd/rbd/util/rbd"
)

const (
	name                = "rbd"
	rate                = 8
	nodeResourceInfoKey = "/resource/rbd/%s"
	priority            = 100
)

// Plugin
type Plugin struct {
	name   string
	config config.Config
	store  meta.KV
	rapi   rbdapi.API
}

// NewPlugin .
func NewPlugin(ctx context.Context, cfg config.Config, t *testing.T) (*Plugin, error) {
	if t == nil && len(cfg.Etcd.Machines) < 1 {
		return nil, coretypes.ErrConfigInvaild
	}
	var err error
	plugin := &Plugin{name: name, config: cfg}
	if plugin.store, err = meta.NewETCD(cfg.Etcd, t); err != nil {
		log.WithFunc("resource.rbd.NewPlugin").Error(ctx, err)
		return nil, err
	}
	if plugin.rapi, err = rbdapi.New(); err != nil {
		log.WithFunc("resource.rbd.NewPlugin").Error(ctx, err)
		return nil, err
	}
	return plugin, nil
}

// Name .
func (p Plugin) Name() string {
	return p.name
}
