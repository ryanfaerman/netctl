package main

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

type asset struct {
	kind     string
	path     string
	name     string
	minified bool
}

func main() {
	var assets []asset

	root := "internal/views/"
	filepath.WalkDir("./"+root, func(path string, d fs.DirEntry, err error) error {
		switch filepath.Ext(path) {
		case ".scss":
			if strings.HasPrefix(d.Name(), "_") {
				assets = append(assets, asset{
					kind:     "scss",
					path:     path,
					name:     filepath.Join(strings.TrimPrefix(filepath.Dir(path), root), strings.TrimPrefix(d.Name(), "_")),
					minified: strings.HasSuffix(d.Name(), ".min.scss"),
				})
			}
		case ".js":
			assets = append(assets, asset{
				kind:     "js",
				path:     path,
				name:     filepath.Join(strings.TrimPrefix(filepath.Dir(path), root), d.Name()),
				minified: strings.HasSuffix(d.Name(), ".min.js"),
			})
		case ".templ", ".go", ".md", ".css":
			return nil

		default:
			if d.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if d.Name() == ext {
				return nil
			}
			assets = append(assets, asset{
				kind:     strings.TrimPrefix(ext, "."),
				path:     path,
				name:     filepath.Join(strings.TrimPrefix(filepath.Dir(path), root), d.Name()),
				minified: strings.HasSuffix(d.Name(), ".min"+ext),
			})
		}
		return nil
	})

	spew.Dump(assets)
}
