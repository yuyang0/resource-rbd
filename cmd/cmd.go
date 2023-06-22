package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/projecteru2/core/log"
	resourcetypes "github.com/projecteru2/core/resource/types"
	zerolog "github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/yuyang0/resource-rbd/config"
	"github.com/yuyang0/resource-rbd/rbd"
	"github.com/yuyang0/resource-rbd/rbd/util/idgen"
)

var (
	ConfigPath      string
	EmbeddedStorage bool
)

func Serve(c *cli.Context, f func(s *rbd.Plugin, in resourcetypes.RawParams) (interface{}, error)) error {
	cfg, err := config.New(ConfigPath)
	if err != nil {
		return cli.Exit(err, 128)
	}

	var t *testing.T
	if EmbeddedStorage {
		t = &testing.T{}
	}

	if err := idgen.Init(cfg.ID); err != nil {
		return cli.Exit(err, 128)
	}
	if err := log.SetupLog(c.Context, cfg.LogLevel, cfg.SentryDSN); err != nil {
		zerolog.Fatal().Err(err).Send()
	}
	s, err := rbd.NewPlugin(c.Context, cfg, t)
	if err != nil {
		return cli.Exit(err, 128)
	}

	in := resourcetypes.RawParams{}
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		return cli.Exit(err, 128)
	}

	if r, err := f(s, in); err != nil {
		return cli.Exit(err, 128)
	} else if o, err := json.Marshal(r); err != nil {
		return cli.Exit(err, 128)
	} else {
		fmt.Print(string(o))
	}
	return nil
}