package rbd

import (
	"context"

	enginetypes "github.com/projecteru2/core/engine/types"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"

	resourcetypes "github.com/projecteru2/core/resource/types"
	rbdtypes "github.com/yuyang0/resource-rbd/rbd/types"
)

// AddNode .
func (p Plugin) AddNode(ctx context.Context, nodename string, resource plugintypes.NodeResourceRequest, info *enginetypes.Info) (resourcetypes.RawParams, error) { //nolint
	return resourcetypes.RawParams{
		"capacity": rbdtypes.NodeResource{},
		"usage":    rbdtypes.NodeResource{},
	}, nil
}

// RemoveNode .
func (p Plugin) RemoveNode(ctx context.Context, nodename string) error { //nolint
	return nil
}

// GetNodesDeployCapacity returns available nodes and total capacity
func (p Plugin) GetNodesDeployCapacity(ctx context.Context, nodenames []string, resource plugintypes.WorkloadResourceRequest) (resourcetypes.RawParams, error) { //nolint
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
	return resourcetypes.RawParams{
		"nodes_deploy_capacity_map": nodesDeployCapacityMap,
		"total":                     total,
	}, nil
}

// SetNodeResourceCapacity sets the amount of total resource info
func (p Plugin) SetNodeResourceCapacity(ctx context.Context, nodename string, resource plugintypes.NodeResource, resourceRequest plugintypes.NodeResourceRequest, delta bool, incr bool) (resourcetypes.RawParams, error) { //nolint
	return resourcetypes.RawParams{
		"before": rbdtypes.NodeResource{},
		"after":  rbdtypes.NodeResource{},
	}, nil
}

// GetNodeResourceInfo .
func (p Plugin) GetNodeResourceInfo(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (resourcetypes.RawParams, error) { //nolint
	return resourcetypes.RawParams{
		"capacity": rbdtypes.NodeResource{},
		"usage":    rbdtypes.NodeResource{},
		"diffs":    nil,
	}, nil
}

// SetNodeResourceInfo .
func (p Plugin) SetNodeResourceInfo(ctx context.Context, nodename string, capacity plugintypes.NodeResource, usage plugintypes.NodeResource) error { //nolint
	return nil
}

// SetNodeResourceUsage .
func (p Plugin) SetNodeResourceUsage(ctx context.Context, nodename string, resource plugintypes.NodeResource, resourceRequest plugintypes.NodeResourceRequest, workloadsResource []plugintypes.WorkloadResource, delta bool, incr bool) (resourcetypes.RawParams, error) { //nolint
	return resourcetypes.RawParams{
		"before": rbdtypes.NodeResource{},
		"after":  rbdtypes.NodeResource{},
	}, nil
}

// GetMostIdleNode .
func (p Plugin) GetMostIdleNode(ctx context.Context, nodenames []string) (resourcetypes.RawParams, error) { //nolint
	var mostIdleNode string
	if len(nodenames) > 0 {
		mostIdleNode = nodenames[0]
	}
	return resourcetypes.RawParams{
		"nodename": mostIdleNode,
		"priority": priority,
	}, nil
}

// FixNodeResource .
func (p Plugin) FixNodeResource(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (resourcetypes.RawParams, error) { //nolint
	return resourcetypes.RawParams{
		"capacity": rbdtypes.NodeResource{},
		"usage":    rbdtypes.NodeResource{},
		"diffs":    nil,
	}, nil
}
