package main

import (
	"context"
	"fmt"
	"os"

	"github.com/projecteru2/core/resource/plugins"
	coretypes "github.com/projecteru2/core/types"

	"github.com/yuyang0/resource-rbd/cmd"
	"github.com/yuyang0/resource-rbd/cmd/calculate"
	"github.com/yuyang0/resource-rbd/cmd/metrics"
	"github.com/yuyang0/resource-rbd/cmd/node"
	"github.com/yuyang0/resource-rbd/cmd/rbd"
	rbdlib "github.com/yuyang0/resource-rbd/rbd"
	"github.com/yuyang0/resource-rbd/version"

	"github.com/urfave/cli/v2"
)

func NewPlugin(ctx context.Context, config coretypes.Config) (plugins.Plugin, error) {
	p, err := rbdlib.NewPlugin(ctx, config, nil)
	return p, err
}

func main() {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(version.String())
	}

	app := cli.NewApp()
	app.Name = version.NAME
	app.Usage = "Run eru resource RBD plugin"
	app.Version = version.VERSION
	app.Commands = []*cli.Command{
		rbd.Name(),
		metrics.Description(),
		metrics.GetMetrics(),

		node.AddNode(),
		node.RemoveNode(),
		node.GetNodesDeployCapacity(),
		node.SetNodeResourceCapacity(),
		node.GetNodeResourceInfo(),
		node.SetNodeResourceInfo(),
		node.SetNodeResourceUsage(),
		node.GetMostIdleNode(),
		node.FixNodeResource(),

		calculate.CalculateDeploy(),
		calculate.CalculateRealloc(),
		calculate.CalculateRemap(),
	}
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Value:       "rbd.yaml",
			Usage:       "config file path for plugin, in yaml",
			Destination: &cmd.ConfigPath,
			EnvVars:     []string{"ERU_RESOURCE_CONFIG_PATH"},
		},
		&cli.BoolFlag{
			Name:        "embedded-storage",
			Usage:       "active embedded storage",
			Destination: &cmd.EmbeddedStorage,
		},
	}
	_ = app.Run(os.Args)
}
