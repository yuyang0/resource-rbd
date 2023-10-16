package rbd

import (
	"context"

	"github.com/mitchellh/mapstructure"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
)

// GetMetricsDescription .
func (p Plugin) GetMetricsDescription(context.Context) (*plugintypes.GetMetricsDescriptionResponse, error) {
	resp := &plugintypes.GetMetricsDescriptionResponse{}
	return resp, mapstructure.Decode([]map[string]any{
		// {
		// 	"name":   "rbd_used",
		// 	"help":   "node used rbd.",
		// 	"type":   "gauge",
		// 	"labels": []string{"podname", "nodename"},
		// },
	}, resp)
}

// GetMetrics .
func (p Plugin) GetMetrics(ctx context.Context, podname, nodename string) (*plugintypes.GetMetricsResponse, error) { //nolint
	// safeNodename := strings.ReplaceAll(nodename, ".", "_")
	metrics := []map[string]any{
		// {
		// 	"name":   "rbd_used",
		// 	"labels": []string{podname, nodename},
		// 	"value":  fmt.Sprintf("%+v", 0),
		// 	"key":    fmt.Sprintf("core.node.%s.rbd.used", safeNodename),
		// },
	}

	resp := &plugintypes.GetMetricsResponse{}
	return resp, mapstructure.Decode(metrics, resp)
}
