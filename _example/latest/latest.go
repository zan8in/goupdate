package main

import (
	"fmt"
	"runtime"

	"github.com/apex/log"
	"github.com/kierdavis/ansi"

	"github.com/zan8in/go-update"
	"github.com/zan8in/go-update/progress"
	"github.com/zan8in/go-update/stores/github"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	ansi.HideCursor()
	defer ansi.ShowCursor()

	// update polls(1) from tj/gh-polls on github
	m := &update.Manager{
		Command: "polls",
		Store: &github.Store{
			Owner:   "tj",
			Repo:    "gh-polls",
			Version: "0.0.3",
		},
	}

	// fetch the new releases
	releases, err := m.LatestReleases()
	if err != nil {
		log.Fatalf("error fetching releases: %s", err)
	}

	// no updates
	if len(releases) == 0 {
		log.Info("no updates")
		return
	}

	// latest release
	latest := releases[0]

	// find the tarball for this system
	a := latest.FindTarball(runtime.GOOS, runtime.GOARCH)
	if a == nil {
		log.Info("no binary for your system")
		return
	}

	// whitespace
	fmt.Println()

	// download tarball to a tmp dir
	tarball, err := a.DownloadProxy(progress.Reader)
	if err != nil {
		log.Fatalf("error downloading: %s", err)
	}

	// install it
	if err := m.Install(tarball); err != nil {
		log.Fatalf("error installing: %s", err)
	}

	fmt.Printf("Updated to %s\n", latest.Version)
}
