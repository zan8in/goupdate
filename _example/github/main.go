package main

import (
	"github.com/zan8in/gologger"
	"github.com/zan8in/goupdate"
	"github.com/zan8in/goupdate/stores/github"
)

func main() {

	owner := "zan8in"
	repo := "afrog"
	version := "2.8.1"

	if result, err := github.Update(owner, repo, version); err != nil {
		gologger.Error().Msg(err.Error())
	} else {
		if result.Status == 2 {
			gologger.Info().Msgf("%s %s", repo, goupdate.LatestVersionTips)
		} else {
			gologger.Info().Msgf("Successfully updated to %s %s\n", repo, result.LatestVersion)
		}
	}

}
