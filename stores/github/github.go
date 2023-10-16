// Package github provides a GitHub release store.
package github

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/google/go-github/github"
	update "github.com/zan8in/goupdate"
	"github.com/zan8in/goupdate/progress"
)

// Store is the store implementation.
type Store struct {
	Owner   string
	Repo    string
	Version string
}

type GithubResult struct {
	Status        int // update success = 1; have latest version = 2;
	LatestVersion string
}

// 更新最新版本
// owner = zan8in
// repo = afrog
// version = 2.8.8 当前版本
func Update(owner, repo, version string) (*GithubResult, error) {
	var (
		result = &GithubResult{}
		err    error
	)

	m := &update.Manager{
		Command: repo + ".exe",
		Store: &Store{
			Owner:   owner,
			Repo:    repo,
			Version: version,
		},
	}

	// fetch the new releases
	releases, err := m.LatestReleases()
	if err != nil {
		return result, fmt.Errorf("error fetching releases: %s", err)
	}

	// no updates
	if len(releases) == 0 {
		result.Status = 2
		return result, nil
	}

	// latest release
	latest := releases[0]

	// find the tarball for this system
	a := latest.FindZip(runtime.GOOS, runtime.GOARCH)
	if a == nil {
		return result, fmt.Errorf("no binary for your system")
	}

	// download tarball to a tmp dir
	tarball, err := a.DownloadProxy(progress.Reader)
	if err != nil {
		return result, fmt.Errorf("error downloading: %s", err)
	}

	// install it
	if err := m.Install(tarball); err != nil {
		return result, fmt.Errorf("error installing: %s", err)
	}

	// gologger.Info().Msgf("Successfully updated to %s %s\n", repo, version)
	result.LatestVersion = version
	result.Status = 1
	return result, nil
}

// GetRelease returns the specified release or ErrNotFound.
func (s *Store) GetRelease(version string) (*update.Release, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gh := github.NewClient(nil)

	r, res, err := gh.Repositories.GetReleaseByTag(ctx, s.Owner, s.Repo, "v"+version)

	if res.StatusCode == 404 {
		return nil, update.ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	return githubRelease(r), nil
}

// LatestReleases returns releases newer than Version, or nil.
func (s *Store) LatestReleases() (latest []*update.Release, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gh := github.NewClient(nil)

	releases, _, err := gh.Repositories.ListReleases(ctx, s.Owner, s.Repo, nil)
	if err != nil {
		return nil, err
	}

	for _, r := range releases {
		tag := r.GetTagName()

		if tag == s.Version || "v"+s.Version == tag {
			break
		}

		latest = append(latest, githubRelease(r))
	}

	return
}

// githubRelease returns a Release.
func githubRelease(r *github.RepositoryRelease) *update.Release {
	out := &update.Release{
		Version:     r.GetTagName(),
		Notes:       r.GetBody(),
		PublishedAt: r.GetPublishedAt().Time,
		URL:         r.GetURL(),
	}

	for _, a := range r.Assets {
		out.Assets = append(out.Assets, &update.Asset{
			Name:      a.GetName(),
			Size:      a.GetSize(),
			URL:       a.GetBrowserDownloadURL(),
			Downloads: a.GetDownloadCount(),
		})
	}

	return out
}
