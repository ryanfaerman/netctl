package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/ryanfaerman/netctl/web/named"
	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
	"gopkg.in/yaml.v2"
)

//go:embed static/*
var staticFS embed.FS

//go:embed bundles/*.yml
var bundleFS embed.FS

func (s *Server) setupStatic() {
	static, _ := fs.Sub(staticFS, "static")

	s.mux.Handle(
		"/static/*",
		http.StripPrefix("/static/", statigz.FileServer(static.(fs.ReadDirFS), brotli.AddEncoding)),
	)

	s.setupStaticRoutes()
}

type StaticRoute struct {
	Template string   `yaml:"template"`
	Routes   []string `yaml:"routes"`
}

type StaticRouteData struct {
	r *http.Request
}

func (srd *StaticRouteData) URLParam(name string) string {
	return chi.URLParam(srd.r, name)
}

func (src *StaticRouteData) URLFor(name string, params ...string) string {
	return named.URLFor(name, params...)
}

type Bootstrap struct {
	Title  string              `yaml:"title"`
	Meta   []map[string]string `yaml:"meta"`
	Link   []map[string]string `yaml:"link"`
	Script []map[string]string `yaml:"script"`
}

type Bundle struct {
	name    string
	path    string
	request *http.Request

	Routes    []string  `yaml:"routes"`
	Template  string    `yaml:"template,omitempty"`
	Bootstrap Bootstrap `yaml:"bootstrap,omitempty"`
}

func (b Bundle) WasmPath() string { return filepath.Join("/static", b.name, b.name+".wasm") }

func (b Bundle) forRequest(r *http.Request) *Bundle {
	b.request = r
	return &b
}

func (b *Bundle) RoutePattern() string {
	return chi.RouteContext(b.request.Context()).RoutePattern()
}

func (s *Server) setupStaticRoutes() {

	bundleFiles, err := bundleFS.ReadDir("bundles")
	if err != nil {
		panic(err.Error())
	}

	bundles := []Bundle{}
	for _, file := range bundleFiles {
		raw, err := bundleFS.ReadFile(filepath.Join("bundles", file.Name()))
		if err != nil {
			panic(err.Error())
		}

		bundle := []Bundle{}
		if err := yaml.Unmarshal(raw, &bundle); err != nil {
			s.logger.Error("unable to parse bundle", "error", err, "bundle", file.Name())
			panic(err.Error())
		}

		for i, _ := range bundle {
			bundle[i].name = strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
			bundle[i].path = filepath.Join("bundles", file.Name())
			if bundle[i].Template != "" {
				bundle[i].Template = filepath.Join("static", bundle[i].name, bundle[i].Template)
			}
		}

		bundles = append(bundles, bundle...)
	}

	templateSources := make(map[string][]string)
	for _, bundle := range bundles {
		unique := true
		for _, t := range templateSources[bundle.name] {
			if t == bundle.Template {
				unique = false
			}
		}
		if bundle.Template == "" {
			continue
		}

		_, err := fs.Stat(staticFS, bundle.Template)
		if err != nil {
			panic(err.Error())
		}

		if unique {
			templateSources[bundle.name] = append(templateSources[bundle.name], bundle.Template)
		}
	}

	templates := make(map[string]*template.Template)
	for n, tmpl := range templateSources {
		templates[n] = template.Must(template.ParseFS(staticFS, tmpl...))
	}

	for _, bundle := range bundles {
		bundle := bundle

		for _, route := range bundle.Routes {
			route := route
			s.logger.Debug(
				"creating bundle handler",
				"route", route,
				"template", bundle.Template,
				"for bundle", bundle.name,
			)
			s.mux.Get(route, func(w http.ResponseWriter, r *http.Request) {
				if bundle.Template == "" {
					bootstrapTemplate.Execute(w, bundle.forRequest(r))
				} else {
					templates[bundle.name].Lookup(filepath.Base(bundle.Template)).Execute(w, bundle.forRequest(r))
				}
			})
		}
	}
}
