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
	nodes := generateNodes(ctx, t, st, 1, 0)
	node := nodes[0]
	var req plugintypes.WorkloadResourceRequest
	var err error

	parse := func(d *plugintypes.CalculateDeployResponse) (eps []*types.EngineParams, wrs []*types.WorkloadResource) {
		assert.NotNil(t, d.EnginesParams)
		assert.NotNil(t, d.WorkloadsResource)
		for _, epRaw := range d.EnginesParams {
			ep := &types.EngineParams{}
			err := ep.Parse(epRaw)
			assert.Nil(t, err)
			eps = append(eps, ep)
		}
		for _, wrRaw := range d.WorkloadsResource {
			wr := &types.WorkloadResource{}
			err := wr.Parse(wrRaw)
			assert.Nil(t, err)
			wrs = append(wrs, wr)
		}
		return
	}
	// normal case
	req = plugintypes.WorkloadResourceRequest{
		"volumes": []string{
			fmt.Sprintf("eru/img0:/dir0:rwm:%v", units.GiB),
			fmt.Sprintf("eru/img1:/dir1:rwm:%v", units.GiB),
		},
	}
	d, err := st.CalculateDeploy(ctx, node, 10, req)
	assert.NoError(t, err)
	assert.NotNil(t, d.EnginesParams)
	eParams, _ := parse(d)
	assert.Len(t, eParams, 10)
	assert.Equal(t, eParams[0].Volumes[0],
		fmt.Sprintf("eru/img0:/dir0:rw:%v", units.GiB))
	assert.Equal(t, eParams[0].Volumes[1],
		fmt.Sprintf("eru/img1:/dir1:rw:%v", units.GiB))
}

func TestCalculateRealloc(t *testing.T) {
	ctx := context.Background()
	st := initRBD(ctx, t)
	nodes := generateNodes(ctx, t, st, 1, 0)
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
	parse := func(d *plugintypes.CalculateReallocResponse) (*types.EngineParams, *types.WorkloadResource, *types.WorkloadResource) {
		assert.NotNil(t, d.EngineParams)
		assert.NotNil(t, d.WorkloadResource)
		ep := &types.EngineParams{}
		err := ep.Parse(d.EngineParams)
		assert.Nil(t, err)

		wr := &types.WorkloadResource{}
		err = wr.Parse(d.WorkloadResource)
		assert.Nil(t, err)

		dwr := &types.WorkloadResource{}
		err = dwr.Parse(d.DeltaResource)
		assert.Nil(t, err)
		return ep, wr, dwr
	}
	// normal case
	// 1. Add one
	req = plugintypes.WorkloadResourceRequest{
		"volume-request": []string{"eru/img1:/dir1:mrw:100GiB"},
	}
	d, err := st.CalculateRealloc(ctx, node, resource, req)
	assert.NoError(t, err)
	eParam, wResource, _ := parse(d)
	assert.False(t, eParam.VolumeChanged)

	assert.Len(t, wResource.Volumes, 3)
	vbs := &types.VolumeBindings{}
	assert.NoError(t, vbs.UnmarshalJSON([]byte(`
	[
		"eru/img1:/dir1:mrw:200GiB",
		"eru/img0:/dir0:rw:100GiB",
		"eru/img2:/dir2:rw:1TB"
	]
	`)))
	assert.Truef(t, vbs.Equal(wResource.Volumes), "===\n%s\n===\n%s\n", litter.Sdump(vbs), litter.Sdump(&wResource.Volumes))

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
	eParam, wResource, _ = parse(d)
	assert.True(t, eParam.VolumeChanged)

	assert.Len(t, wResource.Volumes, 2)
	vbs = &types.VolumeBindings{}
	assert.NoError(t, vbs.UnmarshalJSON([]byte(`
	[
		"eru/img4:/dir4:rw:2TB",
		"eru/img0:/dir0:rw:100GiB"
	]
	`)))
	assert.Truef(t, vbs.Equal(wResource.Volumes), "===\n%s\n===\n%s\n", litter.Sdump(vbs), litter.Sdump(&wResource.Volumes))
}

func TestCalculateRemap(t *testing.T) {
	ctx := context.Background()
	st := initRBD(ctx, t)
	nodes := generateNodes(ctx, t, st, 1, 0)
	node := nodes[0]
	d, err := st.CalculateRemap(ctx, node, nil)
	assert.NoError(t, err)
	assert.Nil(t, d.EngineParamsMap)
}
