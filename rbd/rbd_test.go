package rbd

import (
	"context"
	"fmt"
	"testing"

	coretypes "github.com/projecteru2/core/types"
	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	p := initRBD(context.Background(), t)
	assert.Contains(t, p.Name(), p.name)
}

func initRBD(ctx context.Context, t *testing.T) *Plugin {
	config := coretypes.Config{
		Etcd: coretypes.EtcdConfig{
			Prefix: "/rbd",
		},
	}

	p, err := NewPlugin(ctx, config, t)
	assert.NoError(t, err)
	return p
}

func generateNodes(
	ctx context.Context, t *testing.T, st *Plugin, nums int, startIdx int,
) []string {
	names := []string{}
	for i := startIdx; i < startIdx+nums; i++ {
		name := fmt.Sprintf("test%v", i)
		names = append(names, name)
	}
	return names
}
