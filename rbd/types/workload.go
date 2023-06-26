package types

import (
	"github.com/mitchellh/mapstructure"
	resourcetypes "github.com/projecteru2/core/resource/types"
)

// WorkloadResource indicate RBD workload resource
type WorkloadResource struct {
	Volumes   VolumeBindings `json:"volumes" mapstructure:"volumes"`
	TotalSize int64
}

func NewWorkloadResoure() *WorkloadResource {
	return &WorkloadResource{
		Volumes: VolumeBindings{},
	}
}

func (w *WorkloadResource) Size() int64 {
	if w.TotalSize <= 0 {
		sz := int64(0)
		for _, vb := range w.Volumes {
			sz += vb.SizeInBytes
		}
		w.TotalSize = sz
	}
	return w.TotalSize
}

func (w *WorkloadResource) DeepCopy() *WorkloadResource {
	ans := &WorkloadResource{
		Volumes:   VolumeBindings{},
		TotalSize: w.TotalSize,
	}
	for _, vb := range w.Volumes {
		ans.Volumes = append(ans.Volumes, vb.DeepCopy())
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
	// for k, vb := range w1.Volumes {
	// 	if _, ok := w.Volumes[k]; !ok {
	// 		w.Volumes[k] = vb
	// 		w.TotalSize += vb.SizeInBytes
	// 	}
	// }
}

// WorkloadResourceRaw includes all possible fields passed by eru-core for editing workload
// for request calculation
type WorkloadResourceRequest struct {
	Volumes   VolumeBindings `json:"volumes" mapstructure:"volumes"`
	TotalSize int64
}

func (w *WorkloadResourceRequest) DeepCopy() *WorkloadResourceRequest {
	ans := &WorkloadResourceRequest{
		TotalSize: w.TotalSize,
	}
	for _, vb := range w.Volumes {
		newVB := *vb
		ans.Volumes = append(ans.Volumes, &newVB)
	}
	return ans
}

// Validate .
func (w *WorkloadResourceRequest) Validate() error {
	sz := int64(0)
	for _, vb := range w.Volumes {
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

func (w *WorkloadResourceRequest) Init() {
	sz := int64(0)
	for _, vb := range w.Volumes {
		sz += vb.SizeInBytes
	}
	w.TotalSize = sz
}

// Parse .
func (w *WorkloadResourceRequest) Parse(rawParams resourcetypes.RawParams) (err error) {
	if w.Volumes, err = NewVolumeBindings(rawParams.OneOfStringSlice("volumes", "volume-request", "volumes-request")); err != nil {
		return err
	}
	w.Init()
	return nil
}
