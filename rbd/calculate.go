package rbd

import (
	"context"

	"github.com/projecteru2/core/log"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/sanity-io/litter"
	rbdtypes "github.com/yuyang0/resource-rbd/rbd/types"
)

// CalculateDeploy .
func (p Plugin) CalculateDeploy(
	ctx context.Context, nodename string, deployCount int,
	resourceRequest plugintypes.WorkloadResourceRequest,
) (
	*plugintypes.CalculateDeployResponse, error,
) {
	logger := log.WithFunc("resource.rbd.CalculateDeploy").WithField("node", nodename)
	req := &rbdtypes.WorkloadResourceRequest{}
	if err := req.Parse(resourceRequest); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, err, "invalid resource opts %+v", req)
		return nil, err
	}

	var enginesParams []*rbdtypes.EngineParams
	var workloadsResource []*rbdtypes.WorkloadResource

	for i := 0; i < deployCount; i++ {
		wrkRes := rbdtypes.NewWorkloadResoure()
		eParams := rbdtypes.EngineParams{}
		for _, vb := range req.Volumes {
			vb1 := *vb
			wrkRes.Volumes = append(wrkRes.Volumes, &vb1)
			eParams.Volumes = append(eParams.Volumes, vb1.ToString(true))
		}
		enginesParams = append(enginesParams, &eParams)
		workloadsResource = append(workloadsResource, wrkRes)
	}
	epRaws := make([]resourcetypes.RawParams, 0, len(enginesParams))
	for _, ep := range enginesParams {
		epRaws = append(epRaws, ep.AsRawParams())
	}
	wrRaws := make([]resourcetypes.RawParams, 0, len(workloadsResource))
	for _, wr := range workloadsResource {
		wrRaws = append(wrRaws, wr.AsRawParams())
	}
	return &plugintypes.CalculateDeployResponse{
		EnginesParams:     epRaws,
		WorkloadsResource: wrRaws,
	}, nil
}

// CalculateRealloc .
func (p Plugin) CalculateRealloc(
	ctx context.Context, nodename string,
	resource plugintypes.WorkloadResource,
	resourceRequest plugintypes.WorkloadResourceRequest,
) (
	*plugintypes.CalculateReallocResponse, error,
) {
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

	req = &rbdtypes.WorkloadResourceRequest{
		Volumes: rbdtypes.MergeVolumeBindings(req.Volumes, originResource.Volumes),
	}

	if err := req.Validate(); err != nil {
		logger.Errorf(ctx, err, "invalid resource opts %+v", litter.Sdump(req))
		return nil, err
	}

	targetWorkloadResource := &rbdtypes.WorkloadResource{
		Volumes: req.Volumes,
	}
	originResSet := map[[3]string]any{}
	for _, vb := range originResource.Volumes {
		originResSet[vb.GetMapKey()] = struct{}{}
	}
	engineParams := &rbdtypes.EngineParams{
		Storage:       targetWorkloadResource.Size(),
		VolumeChanged: len(originResSet) != len(targetWorkloadResource.Volumes),
	}
	for _, vb := range targetWorkloadResource.Volumes {
		if _, ok := originResSet[vb.GetMapKey()]; !ok {
			engineParams.VolumeChanged = true
		}
		engineParams.Volumes = append(engineParams.Volumes, vb.ToString(true))
	}
	deltaWorkloadResource := getDeltaWorkloadResourceArgs(originResource, targetWorkloadResource)
	return &plugintypes.CalculateReallocResponse{
		EngineParams:     engineParams.AsRawParams(),
		DeltaResource:    deltaWorkloadResource.AsRawParams(),
		WorkloadResource: targetWorkloadResource.AsRawParams(),
	}, nil
}

// CalculateRemap .
func (p Plugin) CalculateRemap(
	context.Context, string,
	map[string]plugintypes.WorkloadResource,
) (
	*plugintypes.CalculateRemapResponse, error,
) {
	return &plugintypes.CalculateRemapResponse{
		EngineParamsMap: nil,
	}, nil
}

func getDeltaWorkloadResourceArgs(originResource, targetWorkloadResource *rbdtypes.WorkloadResource) *rbdtypes.WorkloadResource {
	ans := rbdtypes.NewWorkloadResoure()
	originSeen := map[[3]string]*rbdtypes.VolumeBinding{}
	for _, vb := range originResource.Volumes {
		originSeen[vb.GetMapKey()] = vb
	}
	for _, vb := range targetWorkloadResource.Volumes {
		newVB := *vb
		if originVB, ok := originSeen[vb.GetMapKey()]; ok {
			newVB.SizeInBytes = vb.SizeInBytes - originVB.SizeInBytes
		}
		ans.Volumes = append(ans.Volumes, &newVB)
	}
	return ans
}
