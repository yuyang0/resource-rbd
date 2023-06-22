package rbd

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	plugintypes "github.com/projecteru2/core/resource/plugins/types"
)

// GetMetricsDescription .
func (p Plugin) GetMetricsDescription(context.Context) (*plugintypes.GetMetricsDescriptionResponse, error) {
	resp := &plugintypes.GetMetricsDescriptionResponse{}
	return resp, mapstructure.Decode([]map[string]interface{}{
		{
			"name":   "rbd_capacity",
			"help":   "node available rbd.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename"},
		},
		{
			"name":   "rbd_used",
			"help":   "node used rbd.",
			"type":   "gauge",
			"labels": []string{"podname", "nodename"},
		},
	}, resp)
}

// GetMetrics .
func (p Plugin) GetMetrics(ctx context.Context, podname, nodename string) (*plugintypes.GetMetricsResponse, error) {
	nodeResourceInfo, err := p.doGetNodeResourceInfo(ctx, nodename)
	if err != nil {
		return nil, err
	}
	safeNodename := strings.ReplaceAll(nodename, ".", "_")
	metrics := []map[string]interface{}{
		{
			"name":   "rbd_capacity",
			"labels": []string{podname, nodename},
			"value":  fmt.Sprintf("%+v", nodeResourceInfo.CapSize()),
			"key":    fmt.Sprintf("core.node.%s.rbd.capacity", safeNodename),
		},
		{
			"name":   "rbd_used",
			"labels": []string{podname, nodename},
			"value":  fmt.Sprintf("%+v", nodeResourceInfo.UsageSize()),
			"key":    fmt.Sprintf("core.node.%s.rbd.used", safeNodename),
		},
	}

	resp := &plugintypes.GetMetricsResponse{}
	return resp, mapstructure.Decode(metrics, resp)
}
