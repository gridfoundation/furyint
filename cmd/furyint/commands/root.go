package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gridfoundation/furyint/config"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
)

var (
	tmconfig  = cfg.DefaultConfig()
	furyconfig = config.DefaultNodeConfig
	logger    = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
)

func init() {
	registerFlagsRootCmd(RootCmd)
}

func registerFlagsRootCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log_level", tmconfig.LogLevel, "log level")
}

// ParseConfig retrieves the default environment configuration,
// sets up the Furyint root and ensures that the root exists
func ParseConfig(cmd *cobra.Command) (*cfg.Config, error) {
	conf := cfg.DefaultConfig()
	err := viper.Unmarshal(conf)
	if err != nil {
		return nil, err
	}

	var home string
	if os.Getenv("FURYINTHOME") != "" {
		home = os.Getenv("FURYINTHOME")
	} else {
		home, err = cmd.Flags().GetString(cli.HomeFlag)
		if err != nil {
			return nil, err
		}
	}

	conf.RootDir = home

	conf.SetRoot(conf.RootDir)
	cfg.EnsureRoot(conf.RootDir)
	if err := conf.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("error in config file: %v", err)
	}

	return conf, nil
}

// RootCmd is the root command for Furyint core.
var RootCmd = &cobra.Command{
	Use:   "furyint",
	Short: "ABCI-client implementation for dYmenion's autonomous rollapps",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		v := viper.GetViper()
		err = v.BindPFlags(cmd.Flags())
		if err != nil {
			return err
		}
		err = furyconfig.GetViperConfig(v)
		if err != nil {
			return err
		}

		tmconfig, err = ParseConfig(cmd)
		if err != nil {
			return err
		}

		if tmconfig.LogFormat == cfg.LogFormatJSON {
			logger = log.NewTMJSONLogger(log.NewSyncWriter(os.Stdout))
		}

		logger, err = tmflags.ParseLogLevel(tmconfig.LogLevel, logger, cfg.DefaultLogLevel)
		if err != nil {
			return err
		}

		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}

		logger = logger.With("module", "main")
		return nil
	},
}
