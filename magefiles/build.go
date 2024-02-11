package main

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ryanfaerman/netctl/magefiles/flags"
	"github.com/ryanfaerman/netctl/magefiles/git"
	"github.com/ryanfaerman/netctl/magefiles/module"
	"github.com/ryanfaerman/netctl/magefiles/target"
	"github.com/ryanfaerman/netctl/workgroup"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	"github.com/charmbracelet/log"
)

var (
	goexe = "go"
	dirs  = []string{"bin", "dist", "tmp"}
)

func ldflags() string {
	f := flags.LDFlags{}

	f["github.com/ryanfaerman/version.ApplicationName"] = filepath.Base(module.Path())
	f["github.com/ryanfaerman/version.BuildDate"] = time.Now().Format(time.RFC3339)
	f["github.com/ryanfaerman/version.BuildTag"] = git.Tag()
	f["github.com/ryanfaerman/version.CommitHash"] = git.CommitHash()

	return f.String()
}

func commands() []string {
	c := []string{}

	if files, err := ioutil.ReadDir("./cmd"); err == nil {
		for _, file := range files {
			if file.IsDir() {
				c = append(c, file.Name())
			}
		}
	}

	return c
}

// FileExists returns whether the given file or directory exists or not
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func size(path string) int64 {
	f, err := os.Stat(path)
	if err != nil {
		return 0
	}

	return f.Size()
}

func ensureDirs() error {
	var b strings.Builder

	for _, dir := range dirs {
		if !exists("./" + dir) {
			fmt.Fprintf(&b, "(+) %s\n", dir)
			if err := os.MkdirAll("./"+dir, 0755); err != nil {
				return err
			}
		} else {
			fmt.Fprintf(&b, "(✓) %s\n", dir)
		}
	}
	log.Info("Ensured output directories", "directories", b.String())
	return nil
}

func init() {
	log.SetPrefix(`🧙‍♂️`)
	log.SetReportTimestamp(false)
	// log.SetReportCaller(true)
	// log.SetLevel(log.DebugLevel)
}

func Build() error {
	log.Info("Building for local use", "state", "starting", "commands", commands())
	totalStart := time.Now()

	mg.SerialDeps(
		ensureDirs,
		Vendor,
		func() { mg.Deps(AssetPipeline, GenerateTempl) },
		Generate,
	)

	for _, command := range commands() {
		binaryPath := filepath.Join("./bin", command)
		sourcePath := filepath.Join(module.Path(), "/cmd", command)

		// lds := ldflags() + `-s -w -extldflags '-static'`
		lds := ldflags() + `-s -w`

		started := time.Now()
		log.Info("Building command", "state", "starting", "command", command, "target", target.Local(), "source", sourcePath)
		if err := sh.Run(goexe, "build",
			"-buildvcs=false",
			"-tags", "osusergo,netgo",
			"-o", binaryPath,
			"-ldflags="+lds,
			"-trimpath", sourcePath,
		); err != nil {
			return err
		}

		log.Info("Building command", "state", "complete", "elapsed", time.Since(started).String(), "command", command, "target", target.Local(), "source", sourcePath)
	}

	log.Info("Building for local use", "state", "complete", "elapsed", time.Since(totalStart).String())

	return nil
}

func Release() error {
	mg.SerialDeps(
		ensureDirs,
		Vendor,
		func() { mg.Deps(AssetPipeline, GenerateTempl) },
		Generate,
	)

	totalStart := time.Now()

	target.Add("linux", "amd64")
	target.Add("linux", "arm64")
	target.Add("linux", "arm")
	target.Add("darwin", "amd64")
	target.Add("darwin", "arm64")

	log.Info("Building for release", "state", "starting", "commands", commands(), "targets", target.All())

	wg := workgroup.New(3)

	lds := ldflags() + `-s -w -extldflags '-static'`

	target.Each(func(t target.Target) {
		wg.Go(func() error {
			if err := wg.Acquire(1); err != nil {
				return err
			}
			defer wg.Release(1)

			for _, command := range commands() {
				binaryPath := filepath.Join("./dist", t.Name())
				sourcePath := filepath.Join(module.Path(), "/cmd", command)

				log.Info("Building command", "state", "starting", "binary", t.Name(), "target", t, "source", sourcePath)
				started := time.Now()
				if err := sh.RunWith(
					t.Env(), goexe,
					"build",
					"-buildvcs=false",
					"-tags", "osusergo,netgo",
					"-o", binaryPath,
					"-ldflags="+lds,
					"-trimpath", sourcePath,
				); err != nil {
					return err
				}
				log.Info("Building command", "state", "starting", "elapsed", time.Since(started).String(), "binary", t.Name(), "target", t, "source", sourcePath)
			}

			return nil
		})
	})

	err := wg.Wait()

	log.Info("Building for release", "state", "complete", "elapsed", time.Since(totalStart).String(), "commands", commands(), "targets", target.All())

	return err
}

func Vendor() {
	log.Info("Dependency management", "state", "starting")
	started := time.Now()
	sh.Run(goexe, "mod", "tidy")
	log.Info("Dependency management", "state", "elapsed", time.Since(started).String(), "complete")
}

func Generate() error {
	log.Info("Code generation", "state", "starting")
	started := time.Now()
	output, err := sh.Output("go", "generate", "./...")
	log.Info("Code generation", "state", "complete", "elapsed", time.Since(started).String(), "results", output)
	return err
}

func GenerateTempl() error {
	log.Info("Generate templ components", "state", "starting")
	started := time.Now()
	{
		output, err := sh.Output("git", "status", "--porcelain")
		if err != nil {
			return err
		}
		if !strings.Contains(output, "templ") {
			log.Warn("Generate templ components", "state", "skipped", "cause", "no changes", "elapsed", time.Since(started).String())
			return nil
		}
	}
	output, err := sh.Output("templ", "generate")
	log.Info("Generate templ components", "state", "complete", "elapsed", time.Since(started).String(), "results", output)
	return err
}

type asset struct {
	kind     string
	path     string
	name     string
	minified bool
}

func AssetPipeline() error {
	log.Info("Asset pipeline", "state", "starting")

	started := time.Now()

	{
		output, err := sh.Output("git", "status", "--porcelain")
		if err != nil {
			return err
		}
		if !strings.Contains(output, "internal/views/styles") && !strings.Contains(output, "internal/views/javascript") && !strings.Contains(output, "internal/views/assets") {
			log.Warn("Asset pipeline", "state", "skipped", "cause", "no changes", "elapsed", time.Since(started).String())
			return nil
		}
	}

	var assets []asset
	steps := []struct {
		name string
		fn   func() (string, error)
	}{
		{
			"Clear temporary files", func() (string, error) {
				return "", sh.Rm("tmp/css")
			},
		},
		{
			"Collect assets", func() (string, error) {
				root := "internal/views/"
				return "", filepath.WalkDir("./"+root, func(path string, d fs.DirEntry, err error) error {
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
			},
		},
		{
			"Run SASS Preprocessor", func() (string, error) {
				errs := []error{}
				outputs := []string{}
				for _, asset := range assets {
					if asset.kind != "scss" {
						continue
					}
					src := asset.path
					dst := filepath.Join("tmp/css", strings.TrimSuffix(asset.name, ".scss")+".css")
					output, err := sh.Output("sass", "--style=compressed", "--no-source-map", src, dst)
					if output != "" {
						outputs = append(outputs, output)
					}
					errs = append(errs, err)
				}
				return strings.Join(outputs, "\n"), errors.Join(errs...)
			},
		},
		{
			"Minify CSS & JS", func() (string, error) {
				errs := []error{}
				outputs := []string{}
				for _, asset := range assets {
					switch asset.kind {
					case "scss":
						src := filepath.Join("tmp/css", strings.TrimSuffix(asset.name, ".scss")+".css")
						dst := filepath.Join("tmp/css", strings.TrimSuffix(asset.name, ".scss")+".min.css")
						output, err := sh.Output("minify", "-q", "-o", dst, src)
						if output != "" {
							outputs = append(outputs, output)
						}
						errs = append(errs, err)
					case "js":
						if asset.minified {
							errs = append(errs, sh.Copy(filepath.Join("tmp/css", asset.name), asset.path))
							break
						}
						src := asset.path
						dst := filepath.Join("tmp/css", strings.TrimSuffix(asset.name, ".js")+".min.js")
						output, err := sh.Output("minify", "-q", "-o", dst, src)
						if output != "" {
							outputs = append(outputs, output)
						}
						errs = append(errs, err)
					case "png", "jpg", "jpeg", "gif", "svg":
						src := asset.path
						dst := filepath.Join("tmp/css", asset.name)
						os.MkdirAll(filepath.Dir(dst), 0755)
						errs = append(errs, sh.Copy(dst, src))

					}
				}
				return strings.Join(outputs, "\n"), errors.Join(errs...)
			},
		},
		{
			"Compress assets", func() (string, error) {
				dirs, err := os.ReadDir("tmp/css")
				if err != nil {
					return "", err
				}
				var errs []error
				var outputs []string
				for _, dir := range dirs {
					if dir.IsDir() {
						files, err := os.ReadDir(filepath.Join("tmp/css", dir.Name()))
						if err != nil {
							errs = append(errs, err)
							break
						}
						for _, file := range files {
							if !strings.Contains(file.Name(), ".min.") && dir.Name() != "assets" {
								continue
							}
							src := filepath.Join("tmp/css", dir.Name(), file.Name())
							output, err := sh.Output("brotli", "./"+src)
							if output != "" {
								outputs = append(outputs, output)
							}
							errs = append(errs, err)

							output, err = sh.Output("zopfli", "./"+src)
							if output != "" {
								outputs = append(outputs, output)
							}
							errs = append(errs, err)
						}
					}
				}

				return strings.Join(outputs, "\n"), errors.Join(errs...)
			},
		},
		// {
		// 	"Stop right here!", func() (string, error) { return "", errors.New("stop the presses") },
		// },
		{
			"Copy Assets to Static Handler", func() (string, error) {
				dirs, err := os.ReadDir("tmp/css")
				if err != nil {
					return "", err
				}

				errs := []error{}
				for _, dir := range dirs {
					if dir.IsDir() {
						files, err := os.ReadDir(filepath.Join("tmp/css", dir.Name()))
						if err != nil {
							errs = append(errs, err)
							break
						}
						for _, file := range files {
							src := filepath.Join("tmp/css", dir.Name(), file.Name())
							dst := filepath.Join("internal/handlers/static", dir.Name(), file.Name())
							os.MkdirAll(filepath.Dir(dst), 0755)
							errs = append(errs, sh.Copy(dst, src))
						}
					}
				}
				return "", errors.Join(errs...)
			},
		},
	}

	var (
		b           strings.Builder
		pipelineErr error
	)

	for _, step := range steps {
		if pipelineErr != nil {
			fmt.Fprintf(&b, "( ) %s\n", step.name)
			continue
		}

		now := time.Now()
		output, err := step.fn()
		elapsed := time.Since(now)
		if err != nil {
			pipelineErr = err
			fmt.Fprintf(&b, "(X) %s in %s\n", err.Error(), elapsed.String())
			continue
		}
		if output != "" {
			fmt.Fprintf(&b, "%s => %s in %s", step.name, output, elapsed.String())
		} else {
			fmt.Fprintf(&b, "(✓) %s in %s\n", step.name, elapsed.String())
		}
	}

	log.Info("Asset pipeline", "state", "complete", "elapsed", time.Since(started).String(), "steps", b.String())
	return pipelineErr
}

func Clean() {
	log.Info("Cleaning output directories", "directories", strings.Join(dirs, "\n"))
	for _, dir := range dirs {
		log.Debug("Removing directory", "directory", dir)
		os.RemoveAll("./" + dir)
	}
}
