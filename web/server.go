package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/ryanfaerman/netctl/config"
	"github.com/ryanfaerman/netctl/web/named"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/pkg/errors"
	"github.com/unrolled/render"
)

type Router interface {
	Routes() chi.Router
}

type Server struct {
	mux    *chi.Mux
	logger *log.Logger
	ln     net.Listener

	servers []*http.Server
	mw      []func(http.Handler) http.Handler
	render  *render.Render
}

type Optioner interface {
	Apply(*Server) error
}

type OptionFunc func(*Server) error

func (of OptionFunc) Apply(s *Server) error {
	return of(s)
}

func WithLogger(l *log.Logger) OptionFunc {
	return func(s *Server) error {
		s.logger = l

		return nil
	}
}

func WithMiddlewares(mw ...func(http.Handler) http.Handler) OptionFunc {
	return func(s *Server) error {
		s.mw = append(s.mw, mw...)

		return nil
	}
}

func NewServer(options ...Optioner) (*Server, error) {
	srv := Server{
		mux: chi.NewRouter(),
		render: render.New(render.Options{
			IndentJSON: true,
		}),
	}

	for _, o := range options {
		if err := o.Apply(&srv); err != nil {
			return nil, err
		}
	}

	srv.mw = append(srv.mw, srv.logging)
	srv.mw = append(srv.mw, Nosurfing)
	srv.mw = append(srv.mw, middleware.Recoverer)
	srv.mw = append(srv.mw, middleware.URLFormat)
	srv.mw = append(srv.mw, middleware.StripSlashes)
	srv.mw = append(srv.mw, WithRender(srv.render))

	for _, mw := range srv.mw {
		srv.mux.Use(mw)
	}

	HookServerRoutes.Dispatch(context.Background(), &srv)

	if config.Flag.Get("web.debug", false) {
		srv.mux.Get("/.well-known/routes", srv.debugRoutes)
	}

	return &srv, nil
}

func (s *Server) debugRoutes(w http.ResponseWriter, r *http.Request) {
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		route = strings.Replace(route, "/*/", "/", -1)
		name := named.Lookup(route)
		if name != "" {
			fmt.Fprintf(w, "%7s %s -> %s\n", method, name, route)
			return nil
		}

		fmt.Fprintf(w, "%7s %s\n", method, route)
		return nil
	}
	if err := chi.Walk(s.mux, walkFunc); err != nil {
		fmt.Printf("Logging err: %s\n", err.Error())
	}
}

func (s *Server) Routes() chi.Router {
	return s.mux
}

// Start the server listening on the given listener. This will not block and
// will continue to service requests untill the listener closes or there is a
// call to Shutdown.
func (s *Server) Start(ln net.Listener) error {
	// s.ln = ln

	srv := &http.Server{
		Addr:    ln.Addr().String(),
		Handler: s.mux,
		BaseContext: func(ln net.Listener) context.Context {
			_, isUnix := ln.(*net.UnixListener)
			return context.WithValue(context.Background(), ctxServerType, isUnix)
		},
	}

	s.servers = append(s.servers, srv)

	go func() {
		l := s.logger.With("addr", srv.Addr)

		l.Info("starting server")

		HookServerStart.Dispatch(context.Background(), ServerStartPayload{*s, ln})

		if err := srv.Serve(ln); err != nil {
			if err == http.ErrServerClosed {
				l.Debug("server is shutdown")
			} else {
				l.Error("unable to start server", "err", err)
			}
		}

	}()

	return nil
}

// Shutdown the server with a timeout from the given context.
func (s *Server) Shutdown(ctx context.Context) error {
	HookServerStop.Dispatch(ctx, ServerStopPayload{ctx})

	for _, srv := range s.servers {
		if err := srv.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "unable to shutdown server gracefully")
		}
	}

	return nil
}

func (s *Server) Stop(ctx context.Context, ln net.Listener) error {
	for _, srv := range s.servers {
		if srv.Addr != ln.Addr().String() {
			continue
		}

		if err := srv.Shutdown(ctx); err != nil {
			return errors.Wrap(err, "unable to shutdown server gracefully")
		}
	}

	return nil
}

func (s *Server) Restart() error { return nil }
