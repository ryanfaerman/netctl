package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryanfaerman/netctl/web"
	"github.com/spf13/cobra"
	webview "github.com/webview/webview_go"
)

var cmdGui = &cobra.Command{
	Use:   "gui",
	Short: "Run the GUI",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		l, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			return err
		}
		defer l.Close()

		fmt.Println("Using port:", l.Addr().(*net.TCPAddr).Port)
		fmt.Println(l.Addr())

		s, err := web.NewServer(web.WithLogger(logger))
		if err != nil {
			return err
		}
		if err := s.Start(l); err != nil {
			logger.Error("could not start", "err", err)
			return err
		}

		w := webview.New(true)
		defer w.Destroy()
		w.SetTitle("netctl")
		w.SetSize(1400, 800, webview.HintNone)
		w.Navigate(fmt.Sprintf("http://%s/v2/net", l.Addr()))

		w.Run()

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

				w.Terminate()

				gracefulTimeout := 15 * time.Second
				select {
				case <-signalCh:
					logger.Info("caught second signal. Exiting", "signal", sig)
					os.Exit(1)
				case <-time.After(gracefulTimeout):
					logger.Error("graceful shutdown timed out. Exiting")
					os.Exit(1)
					logger.Info("graceful exit complete")
					os.Exit(0)
				}
			}
		}

		return nil
	},
}
