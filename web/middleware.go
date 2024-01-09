package web

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/charmbracelet/log"
	"github.com/justinas/nosurf"
	"github.com/unrolled/render"
)

func Nosurfing(h http.Handler) http.Handler {
	surfing := nosurf.New(h)
	surfing.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Warn("failed to validate CSRF token", "reason", nosurf.Reason(r))
		w.WriteHeader(http.StatusBadRequest)
	}))
	return surfing
}

func Dropper(paths ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range paths {
				if path == r.URL.Path {
					http.NotFound(w, r)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

type key int

const (
	ctxRender key = iota
	ctxMountPath
	ctxServerType
	ctxLogger
	ctxSession
)

func SocketOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isUnix, ok := r.Context().Value(ctxServerType).(bool)
		if !ok {
			panic("at the disco!")
		}
		if !isUnix {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func WithRender(ren *render.Render) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxRender, ren)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Respond(r *http.Request) *render.Render {
	rndr, ok := r.Context().Value(ctxRender).(*render.Render)
	if !ok {
		return render.New()
	}
	return rndr
}

type statusRecorder struct {
	http.ResponseWriter
	status   int
	hijacked bool
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	if !sr.hijacked {
		sr.ResponseWriter.WriteHeader(code)

	}
}

func (sr *statusRecorder) Flush() {
	f, ok := sr.ResponseWriter.(http.Flusher)
	if !ok {
		panic("http.ResponseWriter does not implement http.Flusher")
	}
	f.Flush()
}

func (sr *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	fmt.Println("we're hijacking")
	h, ok := sr.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	sr.hijacked = true
	return h.Hijack()
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rw := statusRecorder{w, http.StatusOK, false}
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(&rw, r)
	})
}
