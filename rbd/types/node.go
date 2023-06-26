package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// NodeResource indicate node cpumem resource
type NodeResource struct{}

// Parse .
func (r *NodeResource) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, r)
}

func (r *NodeResource) Validate() error {
	return nil
}

// NodeResourceInfo indicate cpumem capacity and usage
type NodeResourceInfo struct {
	Capacity *NodeResource `json:"capacity"`
	Usage    *NodeResource `json:"usage"`
}

func (n *NodeResourceInfo) Validate() error {
	if err := n.Capacity.Validate(); err != nil {
		return err
	}
	return n.Usage.Validate()
}

// NodeResourceRequest includes all possible fields passed by eru-core for editing node, it not parsed!
type NodeResourceRequest struct {
	SizeInBytes int64 `json:"size_in_bytes" mapstructure:"size_in_bytes"`
}

func (n *NodeResourceRequest) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, n)
}
