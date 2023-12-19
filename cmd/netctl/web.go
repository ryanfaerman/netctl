package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryanfaerman/netctl/health"
	"github.com/ryanfaerman/netctl/hook"

	"github.com/ryanfaerman/netctl/web"
	"github.com/spf13/cobra"
)

var (
	webAddr = ":8090"

	bindErr = errors.New("unable to bind to address")

	cmdWeb = &cobra.Command{
		Use:   "web",
		Short: "Run the web server",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {

			health.Hook.Register(func(e hook.Event[*health.Check]) {
				e.Payload.Add("check.awesome", errors.New("everything is awesome"))

			})
			health.Hook.Register(func(e hook.Event[*health.Check]) {
				e.Payload.Add("check.hotdog", errors.New("not hotdog"))
			})

			s, err := web.NewServer(web.WithLogger(logger))
			if err != nil {
				return err
			}

			l, err := net.Listen("tcp4", webAddr)
			if err != nil {
				return errors.Join(err, bindErr)
			}
			defer l.Close()

			// TODO: remove the socket before starting
			socket, err := net.Listen("unix", "./tmp/retro.sock")
			if err != nil {
				return errors.Join(err, bindErr)
			}
			defer socket.Close()

			if err := s.Start(l); err != nil {
				logger.Error("could not start", "err", err)
				return err
			}

			if err := s.Start(socket); err != nil {
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

func init() {
	cmdWeb.PersistentFlags().StringVarP(&webAddr, "addr", "a", webAddr, "address to service")

}
