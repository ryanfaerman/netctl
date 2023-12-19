package main

import (
	"fmt"
	"strings"

	"github.com/ryanfaerman/netctl/config"

	"github.com/spf13/cobra"
)

var (
	cmdConfig = &cobra.Command{
		Use:   "config",
		Short: "work with system configuration",
		RunE: func(_ *cobra.Command, args []string) error {
			var configs strings.Builder

			for _, item := range config.All() {
				fmt.Fprintf(&configs, "%s => %s\n", item.Uri, item.Data)
			}

			var flags strings.Builder
			for _, item := range config.Flag.All() {
				fmt.Fprintf(&flags, "%s => %s\n", item.Uri, item.Data)
			}

			logger.Info("Config loaded", "config", configs.String(), "flags", flags.String())

			return nil
		},
	}

	cmdConfigGet = &cobra.Command{
		Use:   "get",
		Short: "read a specific config uri",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {

			data := config.Get(args[0])

			fmt.Println(data)

			return nil
		},
	}

	cmdConfigSet = &cobra.Command{
		Use:     "set",
		Short:   "set a specific config uri",
		Aliases: []string{"add"},
		Args:    cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			return config.Set(args[0], args[1])
		},
	}

	cmdConfigUnset = &cobra.Command{
		Use:     "remove",
		Short:   "set a specific config uri",
		Aliases: []string{"unset", "rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return config.Unset(args[0])
		},
	}

	cmdConfigReset = &cobra.Command{
		Use:   "reset",
		Short: "reset all config to default values",
		RunE: func(_ *cobra.Command, args []string) error {
			return config.Reset()
		},
	}

	cmdConfigFlag = &cobra.Command{
		Use:     "flag",
		Short:   "enable a specific flag",
		Aliases: []string{"activate", "enable"},
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return config.Flag.Set(args[0], true)
		},
	}
	cmdConfigUnflag = &cobra.Command{
		Use:     "unflag",
		Short:   "enable a specific flag",
		Aliases: []string{"deactivate", "disable"},
		Args:    cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return config.Flag.Set(args[0], false)
		},
	}
)

func init() {
	cmdConfig.AddCommand(
		cmdConfigGet,
		cmdConfigSet,
		cmdConfigUnset,
		cmdConfigReset,
		cmdConfigFlag,
		cmdConfigUnflag,
	)
}
