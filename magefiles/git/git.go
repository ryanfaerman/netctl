package git

import (
	"os"
	"strings"

	"github.com/ryanfaerman/netctl/magefiles/shell"
)

const (
	UnknownCommitHash = "deadbeef"
	UnknownBranchName = "unknown"
)

func Tag() string {
	var err error
	tagOrBranch, ok := os.LookupEnv("CI_COMMIT_REF_NAME")
	if !ok {
		tagOrBranch, err = shell.Exec("git", "describe", "--tags")
		if err != nil {
			return "dev"
		}
	}

	return strings.TrimSuffix(tagOrBranch, "\n")
}

func CommitHash() string {
	var err error

	hash, ok := os.LookupEnv("CI_COMMIT_SHA")
	if !ok {
		hash, err = shell.Exec("git", "rev-parse", "HEAD")
		if err != nil {
			return UnknownCommitHash
		}
	}

	hash = strings.TrimSpace(hash)
	if len(hash) >= 8 {
		return strings.TrimSpace(hash)[0:8]
	}

	return UnknownCommitHash
}

func Branch() string {
	branch, err := shell.Exec("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return UnknownBranchName
	}
	branch = strings.TrimSpace(branch)
	return branch
}
