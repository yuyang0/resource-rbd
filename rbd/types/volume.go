package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/projecteru2/core/utils"
)

// VolumeBinding format =>  pool/image:dst[:flags][:size][:read_IOPS:write_IOPS:read_bytes:write_bytes]
type VolumeBinding struct {
	Pool        string `json:"pool" mapstructure:"pool"`
	Image       string `json:"image" mapstructure:"image"`
	Destination string `json:"dest" mapstructure:"destination"`
	Flags       string `json:"flags" mapstructure:"flags"`
	SizeInBytes int64  `json:"size_in_bytes" mapstructure:"size_in_bytes"`
	ReadIOPS    int64  `json:"read_iops" mapstructure:"read_iops"`
	WriteIOPS   int64  `json:"write_iops" mapstructure:"write_iops"`
	ReadBPS     int64  `json:"read_bps" mapstructure:"read_bps"`
	WriteBPS    int64  `json:"write_bps" mapstructure:"write_bps"`
}

func (vb *VolumeBinding) GetSource() string {
	return fmt.Sprintf("%s/%s", vb.Pool, vb.Image)
}

func (vb *VolumeBinding) GetMapKey() [3]string {
	return [3]string{vb.Pool, vb.Image, vb.Destination}
}

func (vb *VolumeBinding) DeepCopy() *VolumeBinding {
	return &VolumeBinding{
		Pool:        vb.Pool,
		Image:       vb.Image,
		Destination: vb.Destination,
		Flags:       vb.Flags,
		SizeInBytes: vb.SizeInBytes,
		ReadIOPS:    vb.ReadIOPS,
		WriteIOPS:   vb.WriteIOPS,
		ReadBPS:     vb.ReadBPS,
		WriteBPS:    vb.WriteBPS,
	}
}

// NewVolumeBinding returns pointer of VolumeBinding
func NewVolumeBinding(volume string) (_ *VolumeBinding, err error) {
	var src, dst, flags string
	var size, readIOPS, writeIOPS, readBPS, writeBPS int64

	parts := strings.Split(volume, ":")
	if len(parts) > 8 || len(parts) < 2 {
		return nil, errors.Wrap(ErrInvalidVolume, volume)
	}
	if len(parts) == 2 {
		parts = append(parts, "rw")
	}
	for len(parts) < 8 {
		parts = append(parts, "0")
	}
	src = parts[0]
	dst = parts[1]
	flags = parts[2]

	ptrs := []*int64{&size, &readIOPS, &writeIOPS, &readBPS, &writeBPS}
	for i, ptr := range ptrs {
		value, err := utils.ParseRAMInHuman(parts[i+3])
		if err != nil {
			return nil, err
		}
		*ptr = value
	}

	flagParts := strings.Split(flags, "")
	sort.Strings(flagParts)

	srcParts := strings.Split(src, "/")
	if len(srcParts) != 2 {
		return nil, errors.Wrapf(ErrInvalidVolume, "wrong source format(pool/image): %s", volume)
	}
	pool, image := srcParts[0], srcParts[1]
	vb := &VolumeBinding{
		Pool:        pool,
		Image:       image,
		Destination: dst,
		Flags:       strings.Join(flagParts, ""),
		SizeInBytes: size,
		ReadIOPS:    readIOPS,
		WriteIOPS:   writeIOPS,
		ReadBPS:     readBPS,
		WriteBPS:    writeBPS,
	}

	if vb.Flags == "" {
		vb.Flags = "rw"
	}

	return vb, vb.Validate()
}

// Validate return error if invalid
// Please note: we allow negative value for SizeInBytes,
// because Realloc uses negative value to descrease the size of volume.
func (vb VolumeBinding) Validate() error {
	if vb.Destination == "" {
		return errors.Wrapf(ErrInvalidVolume, "dest must be provided: %+v", vb)
	}
	if vb.Pool == "" || vb.Image == "" {
		return errors.Wrapf(ErrInvalidVolume, "pool and image must be provided: %+v", vb)
	}
	return nil
}

// ToString returns volume string
func (vb VolumeBinding) ToString(normalize bool) (volume string) {
	flags := vb.Flags
	if normalize {
		flags = strings.ReplaceAll(flags, "m", "")
	}

	if strings.Contains(flags, "o") {
		flags = strings.ReplaceAll(flags, "o", "")
		flags = strings.ReplaceAll(flags, "r", "ro")
		flags = strings.ReplaceAll(flags, "w", "wo")
	}
	src := fmt.Sprintf("%s/%s", vb.Pool, vb.Image)
	if !normalize {
		volume = fmt.Sprintf("%s:%s:%s:%d:%d:%d:%d:%d", src, vb.Destination, flags, vb.SizeInBytes, vb.ReadIOPS, vb.WriteIOPS, vb.ReadBPS, vb.WriteBPS)
	} else {
		switch {
		case vb.Flags == "" && vb.SizeInBytes == 0:
			volume = fmt.Sprintf("%s:%s", src, vb.Destination)
		case vb.ReadIOPS != 0 || vb.WriteIOPS != 0 || vb.ReadBPS != 0 || vb.WriteBPS != 0:
			volume = fmt.Sprintf("%s:%s:%s:%d:%d:%d:%d:%d", src, vb.Destination, flags, vb.SizeInBytes, vb.ReadIOPS, vb.WriteIOPS, vb.ReadBPS, vb.WriteBPS)
		default:
			volume = fmt.Sprintf("%s:%s:%s:%d", src, vb.Destination, flags, vb.SizeInBytes)
		}
	}
	return volume
}

type VolumeBindings []*VolumeBinding

func (vbs VolumeBindings) Equal(vbs1 VolumeBindings) bool {
	if len(vbs) != len(vbs1) {
		return false
	}
	seen := map[[3]string]*VolumeBinding{}
	for _, vb := range vbs {
		seen[vb.GetMapKey()] = vb
	}
	for _, vb1 := range vbs1 {
		vb, ok := seen[vb1.GetMapKey()]
		if !ok {
			return false
		}
		if *vb != *vb1 {
			return false
		}
	}
	return true
}
func (vbs VolumeBindings) TotalSize() int64 {
	ans := int64(0)
	for _, vb := range vbs {
		ans += vb.SizeInBytes
	}
	return ans
}

func (vbs *VolumeBindings) UnmarshalJSON(b []byte) (err error) {
	volumes := []string{}
	if err = json.Unmarshal(b, &volumes); err != nil {
		return err
	}
	*vbs, err = NewVolumeBindings(volumes)
	return
}

// MarshalJSON is used for encoding/json.Marshal
func (vbs VolumeBindings) MarshalJSON() ([]byte, error) {
	volumes := []string{}
	for _, vb := range vbs {
		volumes = append(volumes, vb.ToString(false))
	}
	bs, err := json.Marshal(volumes)
	return bs, err
}

// NewVolumeBindings return VolumeBindings of reference type
func NewVolumeBindings(volumes []string) (volumeBindings VolumeBindings, err error) {
	for _, vb := range volumes {
		volumeBinding, err := NewVolumeBinding(vb)
		if err != nil {
			return nil, err
		}
		volumeBindings = append(volumeBindings, volumeBinding)
	}
	return
}

// Validate .
func (vbs VolumeBindings) Validate() error {
	seenDest := map[string]bool{}
	seenSrc := map[string]bool{}
	for _, vb := range vbs {
		if err := vb.Validate(); err != nil {
			return errors.Wrapf(ErrInvalidVolumes, "invalid VolumeBinding: %s", err)
		}
		v := seenDest[vb.Destination]
		if v {
			return errors.Wrapf(ErrInvalidVolumes, "duplicated destination: %s", vb.Destination)
		}
		seenDest[vb.Destination] = true

		src := vb.GetSource()
		if v := seenSrc[src]; v {
			if vb.Image != "" {
				return errors.Wrapf(ErrInvalidVolumes, "duplicated source: %s", src)
			}
		}
		seenSrc[src] = true
	}
	return nil
}

// MergeVolumeBindings combines two VolumeBindings
func MergeVolumeBindings(vbs1 VolumeBindings, vbs2 ...VolumeBindings) (ans VolumeBindings) {
	vbMap := map[[3]string]*VolumeBinding{}
	for _, vbs := range append(vbs2, vbs1) {
		for _, vb := range vbs {
			if binding, ok := vbMap[vb.GetMapKey()]; ok {
				binding.SizeInBytes += vb.SizeInBytes
				binding.ReadIOPS += vb.ReadIOPS
				binding.WriteIOPS += vb.WriteIOPS
				binding.ReadBPS += vb.ReadBPS
				binding.WriteBPS += vb.WriteBPS
			} else {
				vbMap[vb.GetMapKey()] = &VolumeBinding{
					Pool:        vb.Pool,
					Image:       vb.Image,
					Destination: vb.Destination,
					Flags:       vb.Flags,
					SizeInBytes: vb.SizeInBytes,
					ReadIOPS:    vb.ReadIOPS,
					WriteIOPS:   vb.WriteIOPS,
					ReadBPS:     vb.ReadBPS,
					WriteBPS:    vb.WriteBPS,
				}
			}
		}
	}

	for _, vb := range vbMap {
		if vb.SizeInBytes > 0 {
			ans = append(ans, vb)
		}
	}
	return ans
}

func RemoveEmptyVolumeBinding(vbs VolumeBindings) VolumeBindings {
	var ans VolumeBindings
	for _, vb := range vbs {
		if vb.SizeInBytes > 0 {
			ans = append(ans, vb)
		}
	}
	return ans
}
