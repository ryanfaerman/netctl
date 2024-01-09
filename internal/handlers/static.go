package handlers

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

//go:embed static/*
var staticFS embed.FS

func init() {
	registerRoutableFunc(func(r chi.Router) {
		static, _ := fs.Sub(staticFS, "static")
		r.Handle(
			"/static/*",
			http.StripPrefix("/static/", statigz.FileServer(static.(fs.ReadDirFS), brotli.AddEncoding)),
		)
	})
}
