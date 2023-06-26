package rbd

import "github.com/yuyang0/resource-rbd/rbd/types"

func getDeltaWorkloadResourceArgs(originResource, targetWorkloadResource *types.WorkloadResource) *types.WorkloadResource {
	ans := types.NewWorkloadResoure()
	ans.TotalSize = targetWorkloadResource.Size() - originResource.Size()
	originSeen := map[[3]string]*types.VolumeBinding{}
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
