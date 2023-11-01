package types

import (
	"encoding/json"

	resourcetypes "github.com/projecteru2/core/resource/types"
)

// EngineParams .
type EngineParams struct {
	Volumes       []string `json:"volumes" mapstructure:"volumes"`
	VolumeChanged bool     `json:"volume_changed" mapstructure:"volume_changed"` // indicates whether the realloc request includes new volumes
	Storage       int64    `json:"storage" mapstructure:"storage"`
}

func (ep *EngineParams) AsRawParams() resourcetypes.RawParams {
	return resourcetypes.RawParams{
		"volumes":        ep.Volumes,
		"volume_changed": ep.VolumeChanged,
		"storage":        ep.Storage,
	}
}

func (ep *EngineParams) Parse(rawParams resourcetypes.RawParams) (err error) {
	// Have to use json because volume plan use customize marshal
	body, err := json.Marshal(rawParams)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, ep)
}
