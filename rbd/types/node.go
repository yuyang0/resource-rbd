package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// NodeResource indicate node cpumem resource
type NodeResource struct {
	SizeInBytes int64 `json:"size_in_bytes" mapstructure:"size_in_bytes"`
}

func NewNodeResource(sz int64) *NodeResource {
	return &NodeResource{
		SizeInBytes: sz,
	}
}

// Parse .
func (r *NodeResource) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, r)
}

func (r *NodeResource) Validate() error {
	return nil
}

// DeepCopy .
func (r *NodeResource) DeepCopy() *NodeResource {
	return &NodeResource{}
}

// Add .
func (r *NodeResource) Add(r1 *NodeResource) {
	r.SizeInBytes += r1.SizeInBytes
}

// Sub .
func (r *NodeResource) Sub(r1 *NodeResource) {
	r.SizeInBytes -= r1.SizeInBytes
}

// NodeResourceInfo indicate cpumem capacity and usage
type NodeResourceInfo struct {
	Capacity *NodeResource `json:"capacity"`
	Usage    *NodeResource `json:"usage"`
}

func (n *NodeResourceInfo) CapSize() int64 {
	return n.Capacity.SizeInBytes
}

func (n *NodeResourceInfo) UsageSize() int64 {
	return n.Usage.SizeInBytes
}

func (n *NodeResourceInfo) AvailableSize() int64 {
	return n.Capacity.SizeInBytes - n.Usage.SizeInBytes
}

// DeepCopy .
func (n *NodeResourceInfo) DeepCopy() *NodeResourceInfo {
	return &NodeResourceInfo{
		Capacity: n.Capacity.DeepCopy(),
		Usage:    n.Usage.DeepCopy(),
	}
}

func (n *NodeResourceInfo) Validate() error {
	if err := n.Capacity.Validate(); err != nil {
		return err
	}
	return n.Usage.Validate()
}

func (n *NodeResourceInfo) GetAvailableResource() *NodeResource {
	availableResource := n.Capacity.DeepCopy()
	availableResource.Sub(n.Usage)

	return availableResource
}

// NodeResourceRequest includes all possible fields passed by eru-core for editing node, it not parsed!
type NodeResourceRequest struct {
	SizeInBytes int64 `json:"size_in_bytes" mapstructure:"size_in_bytes"`
}

func (n *NodeResourceRequest) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, n)
}
