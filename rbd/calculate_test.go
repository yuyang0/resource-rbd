package rbd

import (
	"context"
	"fmt"
	"testing"

	"github.com/docker/go-units"
	"github.com/mitchellh/mapstructure"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
	"github.com/yuyang0/resource-rbd/rbd/types"
)

func TestCalculateDeploy(t *testing.T) {
	ctx := context.Background()
	st := initRBD(ctx, t)
	vols := []string{"/data0:1T", "/data1:1T", "/data2:1T", "/data3:1T"}
	nodes := generateNodes(ctx, t, st, 1, vols, 0)
	node := nodes[0]
	var req plugintypes.WorkloadResourceRequest
	var err error

	// normal case
	req = plugintypes.WorkloadResourceRequest{
		"volumes": []string{
			fmt.Sprintf("eru/img0:/dir0:rwm:%v", units.GiB),
			fmt.Sprintf("eru/img1:/dir1:rwm:%v", units.GiB),
		},
	}
	d, err := st.CalculateDeploy(ctx, node, 10, req)
	assert.NoError(t, err)
	assert.NotNil(t, d["engines_params"])
	eParams := d["engines_params"].([]*types.EngineParams)
	assert.Len(t, eParams, 10)
	assert.Equal(t, eParams[0].Volumes[0],
		fmt.Sprintf("eru/img0:/dir0:rw:%v", units.GiB))
	assert.Equal(t, eParams[0].Volumes[1],
		fmt.Sprintf("eru/img1:/dir1:rw:%v", units.GiB))
}

func TestCalculateRealloc(t *testing.T) {
	ctx := context.Background()
	st := initRBD(ctx, t)
	vols := []string{"/data0:1T", "/data1:1T", "/data2:1T", "/data3:1T"}
	nodes := generateNodes(ctx, t, st, 1, vols, 0)
	node := nodes[0]

	bindings, err := types.NewVolumeBindings([]string{
		"eru/img0:/dir0:rw:100GiB",
		"eru/img1:/dir1:mrw:100GiB",
		"eru/img2:/dir2:rw:1TB",
	})
	assert.NoError(t, err)

	wrkResource := &types.WorkloadResource{
		Volumes: bindings,
	}
	resource := plugintypes.WorkloadResource{}
	assert.NoError(t, mapstructure.Decode(wrkResource, &resource))

	req := plugintypes.WorkloadResourceRequest{}

	// normal case
	// 1. Add one
	req = plugintypes.WorkloadResourceRequest{
		"volume-request": []string{"eru/img1:/dir1:mrw:100GiB"},
	}
	d, err := st.CalculateRealloc(ctx, node, resource, req)
	assert.NoError(t, err)
	v, ok := d["engine_params"].(*types.EngineParams)
	assert.True(t, ok)
	assert.False(t, v.VolumeChanged)

	v2, ok := d["workload_resource"].(*types.WorkloadResource)
	assert.True(t, ok)
	assert.Len(t, v2.Volumes, 3)
	vbs := &types.VolumeBindings{}
	assert.NoError(t, vbs.UnmarshalJSON([]byte(`
	[
		"eru/img0:/dir0:rw:100GiB",
		"eru/img1:/dir1:mrw:200GiB",
		"eru/img2:/dir2:rw:1TB"
	]
	`)))
	assert.Equal(t, litter.Sdump(vbs), litter.Sdump(&v2.Volumes))

	// 2. delete One
	req = plugintypes.WorkloadResourceRequest{
		"volume-request": []string{
			"eru/img1:/dir1:mrw:-100GiB",
			"eru/img2:/dir2:rw:-2TB",
			"eru/img3:/dir3:rw:-2TB",
			"eru/img4:/dir4:rw:2TB",
		},
	}
	d, err = st.CalculateRealloc(ctx, node, resource, req)
	assert.NoError(t, err)
	v, ok = d["engine_params"].(*types.EngineParams)
	assert.True(t, ok)
	assert.True(t, v.VolumeChanged)

	v2, ok = d["workload_resource"].(*types.WorkloadResource)
	assert.True(t, ok)
	assert.Len(t, v2.Volumes, 2)
	vbs = &types.VolumeBindings{}
	assert.NoError(t, vbs.UnmarshalJSON([]byte(`
	[
		"eru/img0:/dir0:rw:100GiB",
		"eru/img4:/dir4:rw:2TB"
	]
	`)))
	assert.Equal(t, litter.Sdump(vbs), litter.Sdump(&v2.Volumes))
}

func TestCalculateRemap(t *testing.T) {
	ctx := context.Background()
	st := initRBD(ctx, t)
	vols := []string{"/data0:1T", "/data1:1T", "/data2:1T", "/data3:1T"}
	nodes := generateNodes(ctx, t, st, 1, vols, 0)
	node := nodes[0]
	d, err := st.CalculateRemap(ctx, node, nil)
	assert.NoError(t, err)
	assert.Nil(t, d["engine_params_map"])
}
