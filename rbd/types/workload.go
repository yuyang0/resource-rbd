package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// WorkloadResource indicate RBD workload resource
type WorkloadResource struct {
	Volumes   VolumeBindings `json:"volumes" mapstructure:"volumes"`
	totalSize int64
}

func NewWorkloadResoure() *WorkloadResource {
	return &WorkloadResource{
		Volumes: VolumeBindings{},
	}
}

func (w *WorkloadResource) Size() int64 {
	if w.totalSize <= 0 {
		sz := int64(0)
		for _, vb := range w.Volumes {
			sz += vb.SizeInBytes
		}
		w.totalSize = sz
	}
	return w.totalSize
}

func (w *WorkloadResource) DeepCopy() *WorkloadResource {
	ans := &WorkloadResource{
		Volumes:   VolumeBindings{},
		totalSize: w.totalSize,
	}
	for _, vb := range w.Volumes {
		ans.Volumes = append(ans.Volumes, vb.DeepCopy())
	}
	return ans
}

// ParseFromRawParams .
func (w *WorkloadResource) Parse(rawParams resourcetypes.RawParams) error {
	return mapstructure.Decode(rawParams, w)
}

// WorkloadResourceRaw includes all possible fields passed by eru-core for editing workload
// for request calculation
type WorkloadResourceRequest struct {
	Volumes VolumeBindings `json:"volumes" mapstructure:"volumes"`
}

func (w *WorkloadResourceRequest) DeepCopy() *WorkloadResourceRequest {
	ans := &WorkloadResourceRequest{}
	for _, vb := range w.Volumes {
		newVB := *vb
		ans.Volumes = append(ans.Volumes, &newVB)
	}
	return ans
}

// Validate .
func (w *WorkloadResourceRequest) Validate() error {
	return w.Volumes.Validate()
}

// Parse .
func (w *WorkloadResourceRequest) Parse(rawParams resourcetypes.RawParams) (err error) {
	if w.Volumes, err = NewVolumeBindings(rawParams.OneOfStringSlice("volumes", "volume-request", "volumes-request")); err != nil {
		return err
	}
	return nil
}
