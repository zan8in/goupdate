package main

import (
	"fmt"
	"runtime"

	"github.com/apex/log"

	update "github.com/zan8in/goupdate"
	"github.com/zan8in/goupdate/progress"
	"github.com/zan8in/goupdate/stores/github"
)

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	// ansi.HideCursor()
	// defer ansi.ShowCursor()

	// update polls(1) from tj/gh-polls on github
	m := &update.Manager{
		Command: "afrog.exe",
		Store: &github.Store{
			Owner:   "zan8in",
			Repo:    "afrog",
			Version: "2.3.1",
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
	a := latest.FindZip(runtime.GOOS, runtime.GOARCH)
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
