package main

import (
	"fmt"
	"runtime"

	"github.com/zan8in/gologger"
	update "github.com/zan8in/goupdate"
	"github.com/zan8in/goupdate/progress"
	"github.com/zan8in/goupdate/stores/github"
)

func main() {

	// update polls(1) from tj/gh-polls on github
	m := &update.Manager{
		Command: "up",
		Store: &github.Store{
			Owner:   "apex",
			Repo:    "up",
			Version: "0.4.6",
		},
	}

	// fetch the target release
	release, err := m.GetRelease("0.4.5")
	if err != nil {
		gologger.Info().Msgf("error fetching release: %s", err)
		return
	}

	// find the tarball for this system
	a := release.FindTarball(runtime.GOOS, runtime.GOARCH)
	if a == nil {
		gologger.Error().Msg("no binary for your system")
		return
	}

	// whitespace
	fmt.Println()

	// download tarball to a tmp dir
	tarball, err := a.DownloadProxy(progress.Reader)
	if err != nil {
		gologger.Error().Msgf("error downloading: %s", err)
		return
	}

	// install it
	if err := m.Install(tarball); err != nil {
		gologger.Error().Msgf("error installing: %s", err)
		return
	}

	gologger.Info().Msgf("Updated to %s\n", release.Version)
}
