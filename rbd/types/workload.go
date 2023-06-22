package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// WorkloadResource indicate RBD workload resource
type WorkloadResource struct {
	Volumes   map[string]VolumeBinding `json:"volumes" mapstructure:"volumes"`
	TotalSize int64
}

func NewWorkloadResoure() *WorkloadResource {
	return &WorkloadResource{
		Volumes: map[string]VolumeBinding{},
	}
}

func (w *WorkloadResource) Size() int64 {
	return w.TotalSize
}

func (w *WorkloadResource) DeepCopy() *WorkloadResource {
	ans := &WorkloadResource{
		Volumes:   map[string]VolumeBinding{},
		TotalSize: w.TotalSize,
	}
	for k, vb := range w.Volumes {
		ans.Volumes[k] = vb
	}
	return ans
}

// ParseFromRawParams .
func (w *WorkloadResource) Parse(rawParams resourcetypes.RawParams) error {
	if err := mapstructure.Decode(rawParams, w); err != nil {
		return err
	}
	sz := int64(0)
	for _, vb := range w.Volumes {
		sz += vb.SizeInBytes
	}
	w.TotalSize = sz
	return nil
}

// Add .
func (w *WorkloadResource) Add(w1 *WorkloadResource) {
	for k, vb := range w1.Volumes {
		if _, ok := w.Volumes[k]; !ok {
			w.Volumes[k] = vb
			w.TotalSize += vb.SizeInBytes
		}
	}
}

// Sub .
func (w *WorkloadResource) Sub(w1 *WorkloadResource) {
	for k, vb := range w1.Volumes {
		if _, ok := w.Volumes[k]; ok {
			delete(w.Volumes, k)
			w.TotalSize -= vb.SizeInBytes
		}
	}

}

type MergeType int

const (
	MergeAdd MergeType = iota
	MergeSub
	MergeTotol
)

func (mt MergeType) Validate() bool {
	return mt >= MergeAdd && mt <= MergeTotol
}

// WorkloadResourceRaw includes all possible fields passed by eru-core for editing workload
// for request calculation
type WorkloadResourceRequest struct {
	MergeType      MergeType      `json:"merge_type" mapstructure:"merge_type"`
	VolumesRequest VolumeBindings `json:"volumes_request" mapstructure:"volumes_request"`
	TotalSize      int64
}

func (w *WorkloadResourceRequest) DeepCopy() *WorkloadResourceRequest {
	ans := &WorkloadResourceRequest{
		TotalSize: w.TotalSize,
	}
	for _, vb := range w.VolumesRequest {
		newVB := *vb
		ans.VolumesRequest = append(ans.VolumesRequest, &newVB)
	}
	return ans
}

// Validate .
func (w *WorkloadResourceRequest) Validate() error {
	sz := int64(0)
	for _, vb := range w.VolumesRequest {
		if vb.SizeInBytes <= 0 {
			return ErrInvalidVolume
		}
		sz += vb.SizeInBytes
	}
	if sz != w.TotalSize {
		return ErrInvalidVolume
	}
	return nil
}

// Parse .
func (w *WorkloadResourceRequest) Parse(rawParams resourcetypes.RawParams) (err error) {
	if err := mapstructure.Decode(rawParams, w); err != nil {
		return err
	}
	sz := int64(0)
	for _, vb := range w.VolumesRequest {
		sz += vb.SizeInBytes
	}
	w.TotalSize = sz
	return nil
}

func (w *WorkloadResourceRequest) MergeFromResource(r *WorkloadResource, mergeTy MergeType) {
	dest2vb := map[string]*VolumeBinding{}
	for _, vb := range w.VolumesRequest {
		dest2vb[vb.Destination] = vb
	}

	switch mergeTy {
	case MergeAdd:
		for _, vb := range r.Volumes {
			vb1, ok := dest2vb[vb.Destination]
			if ok {
				// if the destination is equal, then it means the rbd already exists,
				// so we only change size, io contraints and etc.
				vb1.Pool = vb.Pool
				vb1.Image = vb.Image
			} else {
				w.VolumesRequest = append(w.VolumesRequest, vb.DeepCopy())
			}
		}
	case MergeSub:
		for _, vb := range r.Volumes {
			delete(dest2vb, vb.Destination)
		}
	case MergeTotol:
		// use request to overwrite resource, so do nothing here
		// if the destination is equal, then it means the rbd already exists,
		// so we only change size, io contraints and etc.
		for _, vb := range r.Volumes {
			if vb1, ok := dest2vb[vb.Destination]; ok {
				vb1.Pool = vb.Pool
				vb1.Image = vb.Image
			}
		}
		return
	}
}
