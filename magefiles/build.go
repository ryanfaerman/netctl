package main

import (
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
	log.Info("Ensuring output directories", "needs", strings.Join(dirs, "\n"))

	created := []string{}
	for _, dir := range dirs {
		if !exists("./" + dir) {
			created = append(created, dir)
			log.Debug("Creating directory", "directory", dir)
			if err := os.MkdirAll("./"+dir, 0755); err != nil {
				return err
			}
		}
	}
	if len(created) > 0 {
		log.Info("Created missing directories", "created", strings.Join(created, "\n"))
	}
	return nil
}

func init() {
	log.SetPrefix(`🧙‍♂️`)
	log.SetReportTimestamp(false)
	// log.SetReportCaller(true)
	// log.SetLevel(log.DebugLevel)
}

func Build() error {
	mg.SerialDeps(ensureDirs, Vendor, Generate, GenerateTempl)

	log.Info("Building for local use")

	for _, command := range commands() {

		binaryPath := filepath.Join("./bin", command)
		sourcePath := filepath.Join(module.Path(), "/cmd", command)

		// lds := ldflags() + `-s -w -extldflags '-static'`
		lds := ldflags() + `-s -w`

		log.Info("Building application", "command", command, "target", target.Local(), "source", sourcePath)
		if err := sh.Run(goexe, "build",
			"-buildvcs=false",
			"-tags", "osusergo,netgo",
			"-o", binaryPath,
			"-ldflags="+lds,
			"-trimpath", sourcePath,
		); err != nil {
			return err
		}
	}

	return nil
}

func Release() error {
	mg.SerialDeps(ensureDirs, Vendor, Generate, GenerateTempl)

	log.Info("Building for release")

	target.Add("linux", "amd64")
	target.Add("linux", "arm64")
	target.Add("linux", "arm")
	target.Add("darwin", "amd64")
	target.Add("darwin", "arm64")

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

				log.Info("Building application", "binary", t.Name(), "target", t, "source", sourcePath)
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
			}

			return nil
		})
	})

	return wg.Wait()

}

func Vendor() {
	log.Info("Updating dependencies")
	sh.Run(goexe, "mod", "tidy")
}

func Generate() error {
	log.Info("Generating components")
	output, err := sh.Output("go", "generate", "./...")
	log.Info("Generated components", "results", output)
	return err
}

func GenerateTempl() error {
	log.Info("Generating templ components")
	output, err := sh.Output("templ", "generate")
	log.Info("Generated templ components", "results", output)
	return err
}

func Clean() {
	log.Info("Cleaning output directories", "directories", strings.Join(dirs, "\n"))
	for _, dir := range dirs {
		log.Debug("Removing directory", "directory", dir)
		os.RemoveAll("./" + dir)
	}
}
