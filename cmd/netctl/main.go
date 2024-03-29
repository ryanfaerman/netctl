package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/health"
	"github.com/ryanfaerman/netctl/hook"

	// "github.com/ryanfaerman/netctl/ui"
	"github.com/ryanfaerman/version"
	"github.com/spf13/cobra"
)

var (
	logger           = log.With("app", "netctl")
	globalLogLevel   = "info"
	globalLogFormat  = "logfmt"
	globalConfigPath = ""

	root = &cobra.Command{
		Use:     "netctl",
		Version: version.String(),
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			switch globalLogFormat {
			case "pretty":
				logger.SetFormatter(log.TextFormatter)
			case "json":
				logger.SetFormatter(log.JSONFormatter)
			case "logfmt":
				logger.SetFormatter(log.LogfmtFormatter)
			}
			lvl, err := log.ParseLevel(globalLogLevel)
			if err != nil {
				logger.Warn("Failed to parse log level, defaulting to info", "error", err, "original", globalLogLevel)
				lvl = log.InfoLevel
			}

			logger.SetLevel(lvl)

			config.Logger = logger
			if globalConfigPath != "" {
				config.LoadFrom(globalConfigPath)
			} else {
				if err := config.Load(); err != nil {
					logger.Fatal("Failed to load config", "error", err)
				}
			}

			health.Logger = logger.With("service", "health")
			hook.Logger = logger.With("service", "hook")

			health.Register()

			// ui.Register()
		},
	}
)

func init() {
	root.PersistentFlags().StringVar(&globalConfigPath, "config", os.Getenv("RETRO_CONFIG_PATH"), "non-standard location of the config database")
	root.PersistentFlags().StringVar(&globalLogLevel, "log-level", "info", "minimum level of logs to print to STDERR")
	root.PersistentFlags().StringVar(&globalLogFormat, "log-format", "text", "show logs as: text, logfmt, json")
	root.AddCommand(
		cmdWeb,
		cmdConfig,
		cmdGui,
		cmdService,
	)
}

func main() {
	root.Execute()
}
