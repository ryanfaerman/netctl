package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryanfaerman/netctl/internal/handlers"
	"github.com/ryanfaerman/netctl/internal/services"
	"github.com/ryanfaerman/netctl/web"
	"github.com/spf13/cobra"
)

var (
	cmdService = &cobra.Command{
		Use:  "service",
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {

			services.SetDatabase(nil)
			handlers.Register()

			s, err := web.NewServer(web.WithLogger(logger))
			if err != nil {
				return err
			}

			l, err := net.Listen("tcp4", webAddr)
			if err != nil {
				return errors.Join(err, bindErr)
			}
			defer l.Close()

			if err := s.Start(l); err != nil {
				logger.Error("could not start", "err", err)
				return err
			}

			signalCh := make(chan os.Signal, 10)
			signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGPIPE)

			for {
				sig := <-signalCh

				switch sig {
				case syscall.SIGHUP:
					logger.Info("caught signal reloading", "signal", sig)

					if err := s.Restart(); err != nil {
						logger.Error("reloading failed", "err", err)
					}

					logger.Info("reload complete")
				default:
					logger.Info("gracefully shutting down", "signal", sig)
					gracefulCh := make(chan struct{})
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						if err := s.Shutdown(ctx); err != nil {
							logger.Error("unable to stop!", "err", err)
							return
						}

						close(gracefulCh)
					}()

					gracefulTimeout := 15 * time.Second
					select {
					case <-signalCh:
						logger.Info("caught second signal. Exiting", "signal", sig)
						os.Exit(1)
					case <-time.After(gracefulTimeout):
						logger.Error("graceful shutdown timed out. Exiting")
						os.Exit(1)
					case <-gracefulCh:
						logger.Info("graceful exit complete")
						os.Exit(0)
					}
				}
			}

			return nil
		},
	}
)
