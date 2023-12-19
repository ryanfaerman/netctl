package cache

import (
	"net/http"
	"strings"
	"time"
)

type ResponseWriter struct {
	http.ResponseWriter

	request *http.Request
}

type CacheControl struct {
	Public bool
	MaxAge time.Duration
}

func (c CacheControl) String() string {
	var b strings.Builder
	if c.Public {
		b.WriteString("public")
	} else {
		b.WriteString("private")
	}

	return b.String()
}

// func parseCacheControl(in string) time.Duration {
// 	if in == "" {
// 		return 0 * time.Second
// 	}
// 	for _, directive := range strings.Split(in, ",") {
// 		directive = strings.ToLower(directive)
// 		directive = strings.Replace(directive, " ", "", -1)
// 		fmt.Println(directive)
// 		tokens := strings.SplitN(directive, "=", 2)

// 	}

// 	return 0 * time.Second
// }

func (w ResponseWriter) Write(d []byte) (int, error) {
	etag := w.ResponseWriter.Header().Get("etag")

	if etag != "" {
		cache[w.request.URL.Path] = d
	}

	return w.ResponseWriter.Write(d)
}

var cache map[string][]byte

func init() {
	cache = make(map[string][]byte)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		etag := r.Header.Get("etag")
		if etag != "" {
			data, ok := cache[r.URL.Path]
			if ok {
				w.Write(data)
				return
			}
		}

		w = ResponseWriter{
			ResponseWriter: w,
			request:        r,
		}

		next.ServeHTTP(w, r)

	})
}
