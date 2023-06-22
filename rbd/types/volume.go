package types

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/projecteru2/core/utils"
)

type VolumeBinding struct {
	Pool        string `json:"pool" mapstructure:"pool"`
	Image       string `json:"image" mapstructure:"image"`
	Destination string `json:"dest" mapstructure:"dest"`
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
	pool, image := "", ""
	switch len(srcParts) {
	case 1:
		pool = srcParts[0]
	case 2:
		pool, image = srcParts[0], srcParts[1]
	default:
		return nil, errors.Wrap(ErrInvalidVolume, volume)
	}
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
func (vb VolumeBinding) Validate() error {
	if vb.Destination == "" {
		return errors.Wrapf(ErrInvalidVolume, "dest must be provided: %+v", vb)
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
