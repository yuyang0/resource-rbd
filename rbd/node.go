package rbd

import (
	"context"

	enginetypes "github.com/projecteru2/core/engine/types"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"

	rbdtypes "github.com/yuyang0/resource-rbd/rbd/types"
)

// AddNode .
func (p Plugin) AddNode(ctx context.Context, nodename string, resource plugintypes.NodeResourceRequest, info *enginetypes.Info) (*plugintypes.AddNodeResponse, error) { //nolint
	capacity := &rbdtypes.NodeResource{}
	usage := &rbdtypes.NodeResource{}
	return &plugintypes.AddNodeResponse{
		Capacity: capacity.AsRawParams(),
		Usage:    usage.AsRawParams(),
	}, nil
}

// RemoveNode .
func (p Plugin) RemoveNode(ctx context.Context, nodename string) (*plugintypes.RemoveNodeResponse, error) { //nolint
	return &plugintypes.RemoveNodeResponse{}, nil
}

// GetNodesDeployCapacity returns available nodes and total capacity
func (p Plugin) GetNodesDeployCapacity(ctx context.Context, nodenames []string, resource plugintypes.WorkloadResourceRequest) (*plugintypes.GetNodesDeployCapacityResponse, error) { //nolint
	nodesDeployCapacityMap := map[string]*plugintypes.NodeDeployCapacity{}
	total := 0

	for _, nodename := range nodenames {
		count := 1000
		nodeDeployCapacity := &plugintypes.NodeDeployCapacity{
			Weight:   1,
			Capacity: count,
			Usage:    0.001,
			Rate:     0.0001,
		}
		total += count
		nodesDeployCapacityMap[nodename] = nodeDeployCapacity
	}
	return &plugintypes.GetNodesDeployCapacityResponse{
		NodeDeployCapacityMap: nodesDeployCapacityMap,
		Total:                 total,
	}, nil
}

// SetNodeResourceCapacity sets the amount of total resource info
func (p Plugin) SetNodeResourceCapacity(ctx context.Context, nodename string, resource plugintypes.NodeResource, resourceRequest plugintypes.NodeResourceRequest, delta bool, incr bool) (*plugintypes.SetNodeResourceCapacityResponse, error) { //nolint
	before := &rbdtypes.NodeResource{}
	after := &rbdtypes.NodeResource{}
	return &plugintypes.SetNodeResourceCapacityResponse{
		Before: before.AsRawParams(),
		After:  after.AsRawParams(),
	}, nil
}

// GetNodeResourceInfo .
func (p Plugin) GetNodeResourceInfo(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (*plugintypes.GetNodeResourceInfoResponse, error) { //nolint
	capacity := &rbdtypes.NodeResource{}
	usage := &rbdtypes.NodeResource{}
	return &plugintypes.GetNodeResourceInfoResponse{
		Capacity: capacity.AsRawParams(),
		Usage:    usage.AsRawParams(),
		Diffs:    nil,
	}, nil
}

// SetNodeResourceInfo .
func (p Plugin) SetNodeResourceInfo(ctx context.Context, nodename string, capacity plugintypes.NodeResource, usage plugintypes.NodeResource) (*plugintypes.SetNodeResourceInfoResponse, error) { //nolint
	return &plugintypes.SetNodeResourceInfoResponse{}, nil
}

// SetNodeResourceUsage .
func (p Plugin) SetNodeResourceUsage(ctx context.Context, nodename string, resource plugintypes.NodeResource, resourceRequest plugintypes.NodeResourceRequest, workloadsResource []plugintypes.WorkloadResource, delta bool, incr bool) (*plugintypes.SetNodeResourceUsageResponse, error) { //nolint
	before := &rbdtypes.NodeResource{}
	after := &rbdtypes.NodeResource{}
	return &plugintypes.SetNodeResourceUsageResponse{
		Before: before.AsRawParams(),
		After:  after.AsRawParams(),
	}, nil
}

// GetMostIdleNode .
func (p Plugin) GetMostIdleNode(ctx context.Context, nodenames []string) (*plugintypes.GetMostIdleNodeResponse, error) { //nolint
	var mostIdleNode string
	if len(nodenames) > 0 {
		mostIdleNode = nodenames[0]
	}
	return &plugintypes.GetMostIdleNodeResponse{
		Nodename: mostIdleNode,
		Priority: priority,
	}, nil
}

// FixNodeResource .
func (p Plugin) FixNodeResource(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (*plugintypes.GetNodeResourceInfoResponse, error) { //nolint
	capacity := &rbdtypes.NodeResource{}
	usage := &rbdtypes.NodeResource{}
	return &plugintypes.GetNodeResourceInfoResponse{
		Capacity: capacity.AsRawParams(),
		Usage:    usage.AsRawParams(),
		Diffs:    nil,
	}, nil
}
