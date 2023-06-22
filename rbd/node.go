package rbd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/cockroachdb/errors"
	enginetypes "github.com/projecteru2/core/engine/types"
	"github.com/projecteru2/core/log"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"

	resourcetypes "github.com/projecteru2/core/resource/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/projecteru2/core/utils"
	"github.com/sanity-io/litter"
	rbdtypes "github.com/yuyang0/resource-rbd/rbd/types"
)

const (
	maxCapacity = math.MaxInt64
)

// AddNode .
func (p Plugin) AddNode(ctx context.Context, nodename string, resource plugintypes.NodeResourceRequest, info *enginetypes.Info) (resourcetypes.RawParams, error) {
	// try to get the node resource
	var err error
	if _, err = p.doGetNodeResourceInfo(ctx, nodename); err == nil {
		return nil, coretypes.ErrNodeExists
	}

	if !errors.IsAny(err, coretypes.ErrInvaildCount, coretypes.ErrNodeNotExists) {
		log.WithFunc("resource.rbd.AddNode").WithField("node", nodename).Error(ctx, err, "failed to get resource info of node")
		return nil, err
	}

	req := &rbdtypes.NodeResourceRequest{}
	if err := req.Parse(resource); err != nil {
		return nil, err
	}
	if req.SizeInBytes == 0 {
		req.SizeInBytes = maxCapacity
	}
	capacity := rbdtypes.NewNodeResource(req.SizeInBytes)
	nodeResourceInfo := &rbdtypes.NodeResourceInfo{
		Capacity: capacity,
		Usage:    rbdtypes.NewNodeResource(0),
	}

	if err = p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		return nil, err
	}
	return resourcetypes.RawParams{
		"capacity": nodeResourceInfo.Capacity,
		"usage":    nodeResourceInfo.Usage,
	}, nil
}

// RemoveNode .
func (p Plugin) RemoveNode(ctx context.Context, nodename string) error {
	var err error
	if _, err = p.store.Delete(ctx, fmt.Sprintf(nodeResourceInfoKey, nodename)); err != nil {
		log.WithFunc("resource.rbd.RemoveNode").WithField("node", nodename).Error(ctx, err, "faield to delete node")
	}
	return err
}

// GetNodesDeployCapacity returns available nodes and total capacity
func (p Plugin) GetNodesDeployCapacity(ctx context.Context, nodenames []string, resource plugintypes.WorkloadResourceRequest) (resourcetypes.RawParams, error) {
	logger := log.WithFunc("resource.rbd.GetNodesDeployCapacity")
	req := &rbdtypes.WorkloadResourceRequest{}
	if err := req.Parse(resource); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, err, "invalid resource opts %+v", req)
		return nil, err
	}

	nodesDeployCapacityMap := map[string]*plugintypes.NodeDeployCapacity{}
	total := 0

	nodesResourceInfos, err := p.doGetNodesResourceInfo(ctx, nodenames)
	if err != nil {
		return nil, err
	}

	for nodename, nodeResourceInfo := range nodesResourceInfos {
		availableResource := nodeResourceInfo.GetAvailableResource()
		count := int(availableResource.SizeInBytes / req.TotalSize)
		nodeDeployCapacity := &plugintypes.NodeDeployCapacity{
			Weight:   1,
			Capacity: count,
			Usage:    float64(nodeResourceInfo.UsageSize()) / float64(nodeResourceInfo.CapSize()),
			Rate:     float64(req.TotalSize) / float64(nodeResourceInfo.CapSize()),
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
func (p Plugin) SetNodeResourceCapacity(ctx context.Context, nodename string, resource plugintypes.NodeResource, resourceRequest plugintypes.NodeResourceRequest, delta bool, incr bool) (resourcetypes.RawParams, error) {
	logger := log.WithFunc("resource.rbd.SetNodeResourceCapacity").WithField("node", "nodename")
	req, nodeResource, _, nodeResourceInfo, err := p.parseNodeResourceInfos(ctx, nodename, resource, resourceRequest, nil)
	if err != nil {
		return nil, err
	}
	origin := nodeResourceInfo.Capacity
	before := origin.DeepCopy()

	if !delta && req != nil {
		req.LoadFromOrigin(origin, resourceRequest)
	}
	nodeResourceInfo.Capacity = p.calculateNodeResource(req, nodeResource, origin, nil, delta, incr)

	if err := p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		logger.Errorf(ctx, err, "node resource info %+v", litter.Sdump(nodeResourceInfo))
		return nil, err
	}

	return resourcetypes.RawParams{
		"before": before,
		"after":  nodeResourceInfo.Capacity,
	}, nil
}

// GetNodeResourceInfo .
func (p Plugin) GetNodeResourceInfo(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (resourcetypes.RawParams, error) {
	nodeResourceInfo, _, diffs, err := p.getNodeResourceInfo(ctx, nodename, workloadsResource)
	if err != nil {
		return nil, err
	}

	return resourcetypes.RawParams{
		"capacity": nodeResourceInfo.Capacity,
		"usage":    nodeResourceInfo.Usage,
		"diffs":    diffs,
	}, nil
}

// SetNodeResourceInfo .
func (p Plugin) SetNodeResourceInfo(ctx context.Context, nodename string, capacity plugintypes.NodeResource, usage plugintypes.NodeResource) error {
	capacityResource := &rbdtypes.NodeResource{}
	usageResource := &rbdtypes.NodeResource{}
	if err := capacityResource.Parse(capacity); err != nil {
		return err
	}
	if err := usageResource.Parse(usage); err != nil {
		return err
	}
	resourceInfo := &rbdtypes.NodeResourceInfo{
		Capacity: capacityResource,
		Usage:    usageResource,
	}

	return p.doSetNodeResourceInfo(ctx, nodename, resourceInfo)
}

// SetNodeResourceUsage .
func (p Plugin) SetNodeResourceUsage(ctx context.Context, nodename string, resource plugintypes.NodeResource, resourceRequest plugintypes.NodeResourceRequest, workloadsResource []plugintypes.WorkloadResource, delta bool, incr bool) (resourcetypes.RawParams, error) {
	logger := log.WithFunc("resource.rbd.SetNodeResourceUsage").WithField("node", "nodename")
	req, nodeResource, wrksResource, nodeResourceInfo, err := p.parseNodeResourceInfos(ctx, nodename, resource, resourceRequest, workloadsResource)
	if err != nil {
		return nil, err
	}
	origin := nodeResourceInfo.Usage
	before := origin.DeepCopy()

	nodeResourceInfo.Usage = p.calculateNodeResource(req, nodeResource, origin, wrksResource, delta, incr)

	if err := p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
		logger.Errorf(ctx, err, "node resource info %+v", litter.Sdump(nodeResourceInfo))
		return nil, err
	}

	return resourcetypes.RawParams{
		"before": before,
		"after":  nodeResourceInfo.Usage,
	}, nil
}

// GetMostIdleNode .
func (p Plugin) GetMostIdleNode(ctx context.Context, nodenames []string) (resourcetypes.RawParams, error) {
	var mostIdleNode string
	var minIdle = math.MaxFloat64

	nodesResourceInfo, err := p.doGetNodesResourceInfo(ctx, nodenames)
	if err != nil {
		return nil, err
	}

	for nodename, nodeResourceInfo := range nodesResourceInfo {
		idle := float64(nodeResourceInfo.UsageSize()) / float64(nodeResourceInfo.CapSize())

		if idle < minIdle {
			mostIdleNode = nodename
			minIdle = idle
		}
	}
	return resourcetypes.RawParams{
		"nodename": mostIdleNode,
		"priority": priority,
	}, nil
}

// FixNodeResource .
func (p Plugin) FixNodeResource(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (resourcetypes.RawParams, error) {
	nodeResourceInfo, actuallyWorkloadsUsage, diffs, err := p.getNodeResourceInfo(ctx, nodename, workloadsResource)
	if err != nil {
		return nil, err
	}

	if len(diffs) != 0 {
		nodeResourceInfo.Usage = &rbdtypes.NodeResource{
			SizeInBytes: actuallyWorkloadsUsage.TotalSize,
		}
		if err = p.doSetNodeResourceInfo(ctx, nodename, nodeResourceInfo); err != nil {
			log.WithFunc("resource.rbd.FixNodeResource").Error(ctx, err)
			diffs = append(diffs, err.Error())
		}
	}
	return resourcetypes.RawParams{
		"capacity": nodeResourceInfo.Capacity,
		"usage":    nodeResourceInfo.Usage,
		"diffs":    diffs,
	}, nil
}

func (p Plugin) getNodeResourceInfo(ctx context.Context, nodename string, workloadsResource []plugintypes.WorkloadResource) (*rbdtypes.NodeResourceInfo, *rbdtypes.WorkloadResource, []string, error) {
	logger := log.WithFunc("resource.rbd.getNodeResourceInfo").WithField("node", nodename)
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		logger.Error(ctx, err)
		return nodeResourceInfo, nil, nil, err
	}

	actuallyWorkloadsUsage := rbdtypes.NewWorkloadResoure()
	for _, workloadResource := range workloadsResource {
		workloadUsage := &rbdtypes.WorkloadResource{}
		if err := workloadUsage.Parse(workloadResource); err != nil {
			logger.Error(ctx, err)
			return nil, nil, nil, err
		}
		actuallyWorkloadsUsage.Add(workloadUsage)
	}

	diffs := []string{}

	if actuallyWorkloadsUsage.Size() != nodeResourceInfo.UsageSize() {
		diffs = append(diffs, fmt.Sprintf("node.VolUsed != sum(workload.VolRequest): %.2d != %.2d", nodeResourceInfo.UsageSize(), actuallyWorkloadsUsage.Size()))
	}
	// for src := range actuallyWorkloadsUsage.Volumes {
	// 	if _, ok := nodeResourceInfo.Usage.GPUMap[addr]; !ok {
	// 		diffs = append(diffs, fmt.Sprintf("%s not in usage", addr))
	// 	}
	// }

	return nodeResourceInfo, actuallyWorkloadsUsage, diffs, nil
}

func (p Plugin) doGetNodeResourceInfo(ctx context.Context, nodename string) (*rbdtypes.NodeResourceInfo, error) {
	key := fmt.Sprintf(nodeResourceInfoKey, nodename)
	resp, err := p.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	r := &rbdtypes.NodeResourceInfo{}
	switch resp.Count {
	case 0:
		return r, errors.Wrapf(coretypes.ErrNodeNotExists, "key: %s", nodename)
	case 1:
		if err := json.Unmarshal(resp.Kvs[0].Value, r); err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, errors.Wrapf(coretypes.ErrInvaildCount, "key: %s", nodename)
	}
}

func (p Plugin) doGetNodesResourceInfo(ctx context.Context, nodenames []string) (map[string]*rbdtypes.NodeResourceInfo, error) {
	keys := []string{}
	for _, nodename := range nodenames {
		keys = append(keys, fmt.Sprintf(nodeResourceInfoKey, nodename))
	}
	resps, err := p.store.GetMulti(ctx, keys)
	if err != nil {
		return nil, err
	}

	result := map[string]*rbdtypes.NodeResourceInfo{}

	for _, resp := range resps {
		r := &rbdtypes.NodeResourceInfo{}
		if err := json.Unmarshal(resp.Value, r); err != nil {
			return nil, err
		}
		result[utils.Tail(string(resp.Key))] = r
	}
	return result, nil
}

func (p Plugin) doSetNodeResourceInfo(ctx context.Context, nodename string, resourceInfo *rbdtypes.NodeResourceInfo) error {
	if err := resourceInfo.Validate(); err != nil {
		return err
	}

	data, err := json.Marshal(resourceInfo)
	if err != nil {
		return err
	}

	_, err = p.store.Put(ctx, fmt.Sprintf(nodeResourceInfoKey, nodename), string(data))
	return err
}

func (p Plugin) overwriteNodeResource(req *rbdtypes.NodeResourceRequest, nodeResource *rbdtypes.NodeResource, origin *rbdtypes.NodeResource, workloadsResource []*rbdtypes.WorkloadResource) *rbdtypes.NodeResource {
	resp := (&rbdtypes.NodeResource{}).DeepCopy() // init nil pointer!
	if req != nil {
		nodeResource = &rbdtypes.NodeResource{
			SizeInBytes: req.SizeInBytes,
		}
	}

	if nodeResource != nil {
		resp.Add(nodeResource)
		return resp
	}

	for _, workloadResource := range workloadsResource {
		nodeResource = &rbdtypes.NodeResource{
			SizeInBytes: workloadResource.TotalSize,
		}
		resp.Add(nodeResource)
	}
	return resp
}

func (p Plugin) incrUpdateNodeResource(req *rbdtypes.NodeResourceRequest, nodeResource *rbdtypes.NodeResource, origin *rbdtypes.NodeResource, workloadsResource []*rbdtypes.WorkloadResource, incr bool) *rbdtypes.NodeResource {
	var resp *rbdtypes.NodeResource
	if origin == nil {
		resp = (&rbdtypes.NodeResource{}).DeepCopy() // init nil pointer!
	} else {
		resp = origin.DeepCopy()
	}

	if req != nil {
		nodeResource = &rbdtypes.NodeResource{
			SizeInBytes: req.SizeInBytes,
		}
	}

	if nodeResource != nil {
		if incr {
			resp.Add(nodeResource)
		} else {
			resp.Sub(nodeResource)
		}
		return resp
	}

	for _, workloadResource := range workloadsResource {
		nodeResource = &rbdtypes.NodeResource{
			SizeInBytes: workloadResource.TotalSize,
		}
		if incr {
			resp.Add(nodeResource)
		} else {
			resp.Sub(nodeResource)
		}
	}
	return resp
}

// calculateNodeResource priority: node resource request > node resource > workload resource args list
func (p Plugin) calculateNodeResource(req *rbdtypes.NodeResourceRequest, nodeResource *rbdtypes.NodeResource, origin *rbdtypes.NodeResource, workloadsResource []*rbdtypes.WorkloadResource, delta bool, incr bool) *rbdtypes.NodeResource {
	if delta {
		//增量更新
		return p.incrUpdateNodeResource(req, nodeResource, origin, workloadsResource, incr)
	} else {
		// 全量更新
		return p.overwriteNodeResource(req, nodeResource, origin, workloadsResource)
	}
}

func (p Plugin) parseNodeResourceInfos(
	ctx context.Context, nodename string,
	resource plugintypes.NodeResource,
	resourceRequest plugintypes.NodeResourceRequest,
	workloadsResource []plugintypes.WorkloadResource,
) (
	*rbdtypes.NodeResourceRequest,
	*rbdtypes.NodeResource,
	[]*rbdtypes.WorkloadResource,
	*rbdtypes.NodeResourceInfo,
	error,
) {
	var req *rbdtypes.NodeResourceRequest
	var nodeResource *rbdtypes.NodeResource
	wrksResource := []*rbdtypes.WorkloadResource{}

	if resourceRequest != nil {
		req = &rbdtypes.NodeResourceRequest{}
		if err := req.Parse(resourceRequest); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	if resource != nil {
		nodeResource = &rbdtypes.NodeResource{}
		if err := nodeResource.Parse(resource); err != nil {
			return nil, nil, nil, nil, err
		}
	}

	for _, workloadResource := range workloadsResource {
		wrkResource := &rbdtypes.WorkloadResource{}
		if err := wrkResource.Parse(workloadResource); err != nil {
			return nil, nil, nil, nil, err
		}
		wrksResource = append(wrksResource, wrkResource)
	}

	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return req, nodeResource, wrksResource, nodeResourceInfo, nil
}
