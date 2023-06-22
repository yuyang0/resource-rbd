package rbd

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/projecteru2/core/log"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
	resourcetypes "github.com/projecteru2/core/resource/types"
	coretypes "github.com/projecteru2/core/types"
	rbdtypes "github.com/yuyang0/resource-rbd/rbd/types"
	"github.com/yuyang0/resource-rbd/rbd/util/idgen"
)

const (
	defaultPool = "eru"
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

	// put resources back into the resource pool
	nodeResourceInfo.Usage.Sub(rbdtypes.NewNodeResource(originResource.TotalSize))

	newReq := req.DeepCopy()
	newReq.MergeFromResource(originResource, req.MergeType)
	if err = newReq.Validate(); err != nil {
		return nil, err
	}

	var enginesParams []*rbdtypes.EngineParams
	var workloadsResource []*rbdtypes.WorkloadResource
	if enginesParams, workloadsResource, err = p.doAlloc(nodeResourceInfo, 1, newReq); err != nil {
		return nil, err
	}

	engineParams := enginesParams[0]
	newResource := workloadsResource[0]

	deltaWorkloadResource := newResource.DeepCopy()
	deltaWorkloadResource.Sub(originResource)

	return resourcetypes.RawParams{
		"engine_params":     engineParams,
		"delta_resource":    deltaWorkloadResource,
		"workload_resource": newResource,
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
	var vbs rbdtypes.VolumeBindings
	var err error

	// rollback if necessary
	defer func() {
		if err == nil {
			return
		}
		for _, vb := range vbs {
			// TODO better way to handle error
			p.rapi.Remove(vb)
		}
	}()
	availableResource := resourceInfo.GetAvailableResource()
	totalSize := int64(0)
	for _, vb := range req.VolumesRequest {
		totalSize += vb.SizeInBytes
	}
	if availableResource.SizeInBytes < int64(deployCount)*totalSize {
		err = coretypes.ErrInsufficientResource
		return enginesParams, workloadsResource, err
	}
	for i := 0; i < deployCount; i++ {
		wrkRes := rbdtypes.NewWorkloadResoure()
		eParams := rbdtypes.EngineParams{}
		for _, vb := range req.VolumesRequest {
			pool := vb.Pool
			image := vb.Image
			if pool == "" {
				pool = defaultPool
			}
			doResize := true
			isExist := false
			if image == "" {
				doResize = false
				image = fmt.Sprintf("img-%s", idgen.GenID())
			}
			vb1 := *vb
			vb1.Pool = pool
			vb1.Image = image
			// create or resize ceph image
			if doResize {
				isExist, err = p.rapi.Exists(&vb1)
				if err != nil {
					return enginesParams, workloadsResource, err
				}
			}
			if doResize && isExist {
				if err = p.rapi.Resize(&vb1); err != nil {
					err = errors.Wrap(err, "failed resize rbd ")
					return enginesParams, workloadsResource, err
				}
			} else {
				if err = p.rapi.Create(&vb1); err != nil {
					err = errors.Wrap(err, "failed create rbd ")
					return enginesParams, workloadsResource, err
				}
				vbs = append(vbs, &vb1)
			}
			wrkRes.Volumes[vb1.GetSource()] = vb1
			eParams.Volumes = append(eParams.Volumes, vb1.ToString(true))
		}
		enginesParams = append(enginesParams, &eParams)
		workloadsResource = append(workloadsResource, wrkRes)
	}
	return enginesParams, workloadsResource, err
}
