package web

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type ctxLogValue struct {
	Path   []string
	Fields *map[string]interface{}
	*sync.Mutex
}

func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := statusRecorder{w, http.StatusOK, false}

		ctx := context.WithValue(r.Context(), ctxLogger, ctxLogValue{
			Fields: &map[string]interface{}{},
			Mutex:  &sync.Mutex{},
		})

		start := time.Now()
		next.ServeHTTP(&rw, r.WithContext(ctx))
		end := time.Now()

		elapsed := end.Sub(start)

		val := ctx.Value(ctxLogger).(ctxLogValue)
		val.Lock()
		defer val.Unlock()

		entry := s.logger.With(
			"http_method", r.Method,
			"http_addr", r.RemoteAddr,
			"http_path", r.URL.Path,
			"http_status", rw.status,
			"http_latency", elapsed,
		)

		for k, v := range *val.Fields {
			entry = entry.With(k, v)
		}

		lfn := entry.Info
		if rw.status >= 500 {
			lfn = entry.Error
		}

		lfn("request completed")

	})
}

// LogWith adds keyvals to the logger's fields.
func LogWith(ctx context.Context, keyvals ...interface{}) {
	val := ctx.Value(ctxLogger).(ctxLogValue)
	val.Lock()
	defer val.Unlock()

	fields := *val.Fields
	for _, s := range val.Path {
		if _, ok := fields[s]; !ok {
			fields[s] = map[string]interface{}{}
		}

		fields = fields[s].(map[string]interface{})
	}

	if len(keyvals)%2 != 0 {
		panic("odd number of keyvals")
	}

	for len(keyvals) >= 2 {
		var (
			key interface{}
			val interface{}
		)
		key, keyvals = keyvals[0], keyvals[1:]
		val, keyvals = keyvals[0], keyvals[1:]
		if keyStr, ok := key.(string); ok {
			fields[keyStr] = val
		}
	}
}
