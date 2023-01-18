package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/indra-labs/indra"
	"github.com/indra-labs/indra/pkg/docker"
	"github.com/indra-labs/indra/pkg/proc/app"
	"github.com/indra-labs/indra/pkg/proc/cmds"
	log2 "github.com/indra-labs/indra/pkg/proc/log"
	"github.com/indra-labs/indra/pkg/proc/opts/config"
	"github.com/indra-labs/indra/pkg/proc/opts/meta"
	"github.com/indra-labs/indra/pkg/proc/opts/toggle"

	"os"
	"time"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

func init() {
	log2.App = "indra"
}

var (
	defaultBuildingTimeout = 800 * time.Second
	defaultRepositoryName  = "indralabs"
	defaultBuildContainer  = "golang:1.19.4"
)

func strPtr(str string) *string { return &str }

var sourceConfigurations = []docker.BuildConfiguration{
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "btcd-source",
		ContextFilePath: "/tmp/btcd-source.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/btcd/intermediate/source/official.Dockerfile",
			Tags: []string{
				"v0.23.3",
			},
			BuildArgs: map[string]*string{
				"sourcing_image":            strPtr(defaultBuildContainer),
				"source_release_url_prefix": strPtr("https://github.com/btcsuite/btcd/releases/download"),
				"source_version":            strPtr("v0.23.3"),
			},
			SuppressOutput: false,
			Remove:         false,
			ForceRemove:    false,
			PullParent:     false,
		},
	},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "lnd-source",
		ContextFilePath: "/tmp/lnd-source.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/lnd/intermediate/source/official.Dockerfile",
			Tags: []string{
				"v0.15.5-beta",
			},
			BuildArgs: map[string]*string{
				"sourcing_image":            strPtr(defaultBuildContainer),
				"source_release_url_prefix": strPtr("https://github.com/lightningnetwork/lnd/releases/download"),
				"source_version":            strPtr("v0.15.5-beta"),
			},
			SuppressOutput: false,
			Remove:         false,
			ForceRemove:    false,
			PullParent:     false,
		},
	},
}

var buildConfigurations = []docker.BuildConfiguration{
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "btcd",
		ContextFilePath: "/tmp/btcd.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/btcd/btcd.Dockerfile",
			Tags: []string{
				"v0.23.3",
				"latest",
			},
			BuildArgs: map[string]*string{
				"source_version":     strPtr("v0.23.3"),
				"scratch_version":    strPtr("latest"),
				"target_os":          strPtr("linux"),
				"target_arch":        strPtr("amd64"),
				"target_arm_version": strPtr(""),
			},
			SuppressOutput: false,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     false,
		},
	},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "btcctl",
		ContextFilePath: "/tmp/btcctl.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/btcd/btcctl.Dockerfile",
			Tags: []string{
				"v0.23.3",
				"latest",
			},
			BuildArgs: map[string]*string{
				"source_version":     strPtr("v0.23.3"),
				"scratch_version":    strPtr("latest"),
				"target_os":          strPtr("linux"),
				"target_arch":        strPtr("amd64"),
				"target_arm_version": strPtr(""),
			},
			SuppressOutput: false,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     false,
		},
	},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "lnd",
		ContextFilePath: "/tmp/lnd.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/lnd/lnd.Dockerfile",
			Tags: []string{
				"v0.15.5-beta",
				"latest",
			},
			BuildArgs: map[string]*string{
				"source_version":     strPtr("v0.15.5-beta"),
				"scratch_version":    strPtr("latest"),
				"target_os":          strPtr("linux"),
				"target_arch":        strPtr("amd64"),
				"target_arm_version": strPtr(""),
			},
			SuppressOutput: false,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     false,
		},
	},
	docker.BuildConfiguration{
		Name:            defaultRepositoryName + "/" + "lncli",
		ContextFilePath: "/tmp/lncli.tar",
		BuildOpts: types.ImageBuildOptions{
			Dockerfile: "docker/lnd/lncli.Dockerfile",
			Tags: []string{
				"v0.15.5-beta",
				"latest",
			},
			BuildArgs: map[string]*string{
				"source_version":     strPtr("v0.15.5-beta"),
				"scratch_version":    strPtr("latest"),
				"target_os":          strPtr("linux"),
				"target_arch":        strPtr("amd64"),
				"target_arm_version": strPtr(""),
			},
			SuppressOutput: false,
			Remove:         true,
			ForceRemove:    true,
			PullParent:     false,
		},
	},
	//docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "lnd",
	//	ContextFilePath: "/tmp/lnd.tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/lnd/Dockerfile",
	//		Tags: []string{
	//			"v0.15.5-beta",
	//			"latest",
	//		},
	//		BuildArgs: map[string]*string{
	//			"base_image":   strPtr(defaultBuildContainer),
	//			"target_image": strPtr("indralabs/scratch:latest"),
	//			// This argument is the tag fetched by git
	//			// It MUST be updated alongside the tag above
	//			"git_repository": strPtr("github.com/lightningnetwork/lnd"),
	//			"git_tag":        strPtr("v0.15.5-beta"),
	//		},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     true,
	//	},
	//},
	// docker.BuildConfiguration{
	//	Name:            defaultRepositoryName + "/" + "indra",
	//	ContextFilePath: "/tmp/indra-" + indra.SemVer + ".tar",
	//	BuildOpts: types.ImageBuildOptions{
	//		Dockerfile: "docker/indra/Dockerfile",
	//		Tags: []string{
	//			indra.SemVer,
	//			"latest",
	//		},
	//		BuildArgs:      map[string]*string{},
	//		SuppressOutput: false,
	//		Remove:         true,
	//		ForceRemove:    true,
	//		PullParent:     true,
	//	},
	// },
}

var commands = &cmds.Command{
	Name:          "release",
	Description:   "Builds the indra docker image and pushes it to a list of docker repositories.",
	Documentation: lorem,
	Default:       cmds.Tags("release"),
	Configs: config.Opts{
		"stable": toggle.New(meta.Data{
			Label:         "stable",
			Description:   "tag the current build as stable.",
			Documentation: lorem,
			Default:       "false",
		}),
		"push": toggle.New(meta.Data{
			Label:         "push",
			Description:   "push the newly built/tagged images to the docker repositories.",
			Documentation: lorem,
			Default:       "false",
		}),
	},
	Entrypoint: func(command *cmds.Command, args []string) error {

		// If we've flagged stable, we should also build a stable tag
		if command.GetValue("stable").Bool() {
			docker.SetRelease()
		}

		// If we've flagged push, the tags will be pushed to all repositories.
		if command.GetValue("push").Bool() {
			docker.SetPush()
		}

		// Set a Timeout for 120 seconds
		ctx, cancel := context.WithTimeout(context.Background(), defaultBuildingTimeout)
		defer cancel()

		// Setup a new instance of the docker client

		var err error
		var cli *client.Client

		if cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation()); check(err) {
			return err
		}

		defer cli.Close()

		// Get ready to submit the builds
		var builder = docker.NewBuilder(ctx, cli, sourceConfigurations, buildConfigurations)

		if err = builder.Build(); check(err) {
			return err
		}

		if err = builder.Push(); check(err) {
			return err
		}

		if err = builder.Close(); check(err) {
			return err
		}

		return nil
	},
}

func main() {

	var err error
	var application *app.App

	// Creates a new application
	if application, err = app.New(commands, os.Args); check(err) {
		os.Exit(1)
	}

	// Launches the application
	if err = application.Launch(); check(err) {
		os.Exit(1)
	}

	os.Exit(0)
}

const lorem = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis 
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. 
Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu 
fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in 
culpa qui officia deserunt mollit anim id est laborum.`
