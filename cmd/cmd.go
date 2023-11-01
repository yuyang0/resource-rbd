package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/projecteru2/core/log"
	resourcetypes "github.com/projecteru2/core/resource/types"
	"github.com/projecteru2/core/utils"
	zerolog "github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/yuyang0/resource-rbd/rbd"
)

var (
	ConfigPath      string
	EmbeddedStorage bool
)

func Serve(c *cli.Context, f func(s *rbd.Plugin, in resourcetypes.RawParams) (interface{}, error)) error {
	cfg, err := utils.LoadConfig(ConfigPath)
	if err != nil {
		return cli.Exit(err, 128)
	}

	var t *testing.T
	if EmbeddedStorage {
		t = &testing.T{}
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
		fmt.Fprintf(os.Stderr, "RBD: failed decode input json: %s\n", err)
		fmt.Fprintf(os.Stderr, "RBD: input: %v\n", in)
		return cli.Exit(err, 128)
	}

	if r, err := f(s, in); err != nil {
		fmt.Fprintf(os.Stderr, "RBD: failed call function: %s\n", err)
		fmt.Fprintf(os.Stderr, "RBD: input: %v\n", in)
		return cli.Exit(err, 128)
	} else if o, err := json.Marshal(r); err != nil {
		fmt.Fprintf(os.Stderr, "RBD: failed encode return object: %s\n", err)
		fmt.Fprintf(os.Stderr, "RBD: input: %v\n", in)
		fmt.Fprintf(os.Stderr, "RBD: output: %v\n", o)
		return cli.Exit(err, 128)
	} else { //nolint
		fmt.Print(string(o))
	}
	return nil
}
