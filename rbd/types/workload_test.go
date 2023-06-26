package types

import (
	"testing"

	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/stretchr/testify/assert"
)

func TestWorkloadResource(t *testing.T) {
	wr := &WorkloadResource{}
	err := wr.Parse(nil)
	assert.Nil(t, err)
}

func TestWorkloadResourceRequest(t *testing.T) {
	// empty request
	req := &WorkloadResourceRequest{}
	err := req.Parse(nil)
	assert.Nil(t, err)
	assert.Nil(t, req.Validate())

	// invalid request
	// 1. duplicate source
	params := resourcetypes.RawParams{
		"volumes": []string{
			"eru/img1:/dir1:mrw:-100GiB",
			"eru/img1:/dir2:rw:-2TB",
		},
	}
	req = &WorkloadResourceRequest{}
	err = req.Parse(params)
	assert.Error(t, req.Validate())

	// 2. duplicate destination
	params = resourcetypes.RawParams{
		"volumes": []string{
			"eru/img1:/dir1:mrw:-100GiB",
			"eru/img2:/dir1:rw:-2TB",
		},
	}
	req = &WorkloadResourceRequest{}
	err = req.Parse(params)
	assert.Error(t, req.Validate())
}
