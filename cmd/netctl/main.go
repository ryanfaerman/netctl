package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/ryanfaerman/netctl/app"
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/health"
	"github.com/ryanfaerman/netctl/hook"
	"github.com/ryanfaerman/netctl/ui"
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

			logger.SetLevel(log.ParseLevel(globalLogLevel))

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
			app.Logger = logger.With("service", "board")

			health.Register()
			app.Register()

			ui.Register()

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
	)
}

func main() {
	root.Execute()
}
