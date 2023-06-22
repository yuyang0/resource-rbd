package rbd

import (
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/urfave/cli/v2"
	"github.com/yuyang0/resource-rbd/cmd"
	"github.com/yuyang0/resource-rbd/rbd"
)

func Name() *cli.Command {
	return &cli.Command{
		Name:   "name",
		Usage:  "show name",
		Action: name,
	}
}

func name(c *cli.Context) error {
	return cmd.Serve(c, func(s *rbd.Plugin, _ resourcetypes.RawParams) (interface{}, error) {
		return s.Name(), nil
	})
}
