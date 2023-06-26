package rbd

import (
	"context"

	"github.com/projecteru2/core/log"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	resourcetypes "github.com/projecteru2/core/resource/types"
	coretypes "github.com/projecteru2/core/types"
	"github.com/sanity-io/litter"
	rbdtypes "github.com/yuyang0/resource-rbd/rbd/types"
)

// CalculateDeploy .
func (p Plugin) CalculateDeploy(ctx context.Context, nodename string, deployCount int, resourceRequest plugintypes.WorkloadResourceRequest) (resourcetypes.RawParams, error) {
	logger := log.WithFunc("resource.rbd.CalculateDeploy").WithField("node", nodename)
	req := &rbdtypes.WorkloadResourceRequest{}
	if err := req.Parse(resourceRequest); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, err, "invalid resource opts %+v", req)
		return nil, err
	}

	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		logger.WithField("node", nodename).Error(ctx, err)
		return nil, err
	}

	var enginesParams []*rbdtypes.EngineParams
	var workloadsResource []*rbdtypes.WorkloadResource

	enginesParams, workloadsResource, err = p.doAlloc(nodeResourceInfo, deployCount, req)
	if err != nil {
		return nil, err
	}

	return resourcetypes.RawParams{
		"engines_params":     enginesParams,
		"workloads_resource": workloadsResource,
	}, nil
}

// CalculateRealloc .
func (p Plugin) CalculateRealloc(ctx context.Context, nodename string, resource plugintypes.WorkloadResource, resourceRequest plugintypes.WorkloadResourceRequest) (resourcetypes.RawParams, error) {
	logger := log.WithFunc("resource.rbd.CalculateRealloc").WithField("node", nodename)
	req := &rbdtypes.WorkloadResourceRequest{}
	if err := req.Parse(resourceRequest); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	originResource := &rbdtypes.WorkloadResource{}
	if err := originResource.Parse(resource); err != nil {
		return nil, err
	}

	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		log.WithFunc("resource.rbd.CalculateRealloc").WithField("node", nodename).Error(ctx, err, "failed to get resource info of node")
		return nil, err
	}
	req = &rbdtypes.WorkloadResourceRequest{
		Volumes: rbdtypes.MergeVolumeBindings(req.Volumes, originResource.Volumes),
	}
	// update size field
	req.Init()

	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, err, "invalid resource opts %+v", litter.Sdump(req))
		return nil, err
	}

	targetWorkloadResource := &rbdtypes.WorkloadResource{
		Volumes: req.Volumes,
	}

	if targetWorkloadResource.Size()-originResource.Size() > nodeResourceInfo.AvailableSize() {
		return nil, coretypes.ErrInsufficientResource
	}

	engineParams := &rbdtypes.EngineParams{}
	for _, vb := range targetWorkloadResource.Volumes {
		engineParams.Volumes = append(engineParams.Volumes, vb.ToString(true))
	}
	deltaWorkloadResource := getDeltaWorkloadResourceArgs(originResource, targetWorkloadResource)
	return resourcetypes.RawParams{
		"engine_params":     engineParams,
		"delta_resource":    deltaWorkloadResource,
		"workload_resource": targetWorkloadResource,
	}, nil
}

// CalculateRemap .
func (p Plugin) CalculateRemap(context.Context, string, map[string]plugintypes.WorkloadResource) (resourcetypes.RawParams, error) {
	return resourcetypes.RawParams{
		"engine_params_map": nil,
	}, nil
}

func (p Plugin) doAlloc(resourceInfo *rbdtypes.NodeResourceInfo, deployCount int, req *rbdtypes.WorkloadResourceRequest) ([]*rbdtypes.EngineParams, []*rbdtypes.WorkloadResource, error) {
	enginesParams := []*rbdtypes.EngineParams{}
	workloadsResource := []*rbdtypes.WorkloadResource{}
	var err error

	availableResource := resourceInfo.GetAvailableResource()
	totalSize := int64(0)
	for _, vb := range req.Volumes {
		totalSize += vb.SizeInBytes
	}
	if availableResource.SizeInBytes < int64(deployCount)*totalSize {
		err = coretypes.ErrInsufficientResource
		return enginesParams, workloadsResource, err
	}
	for i := 0; i < deployCount; i++ {
		wrkRes := rbdtypes.NewWorkloadResoure()
		eParams := rbdtypes.EngineParams{}
		for _, vb := range req.Volumes {
			vb1 := *vb
			wrkRes.Volumes = append(wrkRes.Volumes, &vb1)
			eParams.Volumes = append(eParams.Volumes, vb1.ToString(false))
		}
		enginesParams = append(enginesParams, &eParams)
		workloadsResource = append(workloadsResource, wrkRes)
	}
	return enginesParams, workloadsResource, err
}
