// Package indra is the root level package for Indranet, a low latency, 
// Lightning Network monetised distributed VPN protocol designed for providing
// strong anonymity to valuable internet traffic.
package indra

import (
	"fmt"
)

var (
	// URL is the git URL for the repository.
	URL = "github.com/Indra-Labs/indra"
	// GitRef is the gitref, as in refs/heads/branchname.
	GitRef = "refs/heads/main"
	// ParentGitCommit is the commit hash of the parent HEAD.
	ParentGitCommit = "eb6e0475f2c0be672155eb1c0ca752ae43a57ecf"
	// BuildTime stores the time when the current binary was built.
	BuildTime = "2022-11-19T10:44:07+01:00"
	// SemVer lists the (latest) git tag on the build.
	SemVer = "v0.0.113"
	// PathBase is the path base returned from runtime caller.
	PathBase = "/home/loki/src/github.com/Indra-Labs/indra/"
	// Major is the major number from the tag.
	Major = 0
	// Minor is the minor number from the tag.
	Minor = 0
	// Patch is the patch version number from the tag.
	Patch = 113
)

// Version returns a pretty printed version information string.
func Version() string {
	return fmt.Sprint(
		"\nRepository Information\n",
		"\tGit repository: "+URL+"\n",
		"\tBranch: "+GitRef+"\n",
		"\tParentGitCommit: "+ParentGitCommit+"\n",
		"\tBuilt: "+BuildTime+"\n",
		"\tSemVer: "+SemVer+"\n",
	)
}
