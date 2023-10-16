package gitee

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/mamh-mixed/go-gitee/gitee"
	update "github.com/zan8in/goupdate"
	"github.com/zan8in/goupdate/progress"
	"golang.org/x/oauth2"
)

type Store struct {
	Owner   string
	Repo    string
	Version string
}

type GiteeResult struct {
	Status        int // update success = 1; have latest version = 2;
	LatestVersion string
}

// 更新最新版本
// owner = zan8in
// repo = afrog
// version = 2.8.8 当前版本
func Update(owner, repo, version string) (*GiteeResult, error) {
	var (
		result = &GiteeResult{}
		err    error
	)

	var command string
	switch runtime.GOOS {
	case "windows":
		command = repo + ".exe"
	default:
		command = repo
	}

	s := &Store{
		Owner:   owner,
		Repo:    repo,
		Version: version,
	}

	m := update.Manager{
		Command: command,
	}

	asset, err := s.LatestGitReleases()
	if err != nil {
		if strings.Contains(err.Error(), update.LatestVersionTips) {
			result.LatestVersion = asset.LatestVersion
			result.Status = 2
			return result, nil
		}
		return nil, err
	}

	// download tarball to a tmp dir
	tarball, err := asset.DownloadProxy(progress.Reader)
	if err != nil {
		return nil, fmt.Errorf("error downloading: %s", err)
	}

	// install it
	if err := m.Install(tarball); err != nil {
		return nil, fmt.Errorf("error installing: %s", err)
	}

	// gologger.Info().Msgf("Successfully updated to %s %s\n", command, s.Version)
	result.LatestVersion = asset.LatestVersion
	result.Status = 1
	return result, nil
}

// LatestReleases returns releases newer than Version, or nil.
func (s *Store) LatestGitReleases() (up *update.Asset, err error) {

	token := os.Getenv("")

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	tc := oauth2.NewClient(ctx, ts)
	client := gitee.NewClient(tc)

	owner := s.Owner
	repo := s.Repo

	var currentOS string
	switch runtime.GOOS {
	case "darwin":
		currentOS = "macOS"
	default:
		currentOS = runtime.GOOS
	}

	releases, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	version := (*releases.TagName)[1:]

	if s.Version == version {
		return &update.Asset{
			LatestVersion: version,
		}, fmt.Errorf("%s %s", s.Repo, update.LatestVersionTips)
	}

	zipName := fmt.Sprintf("%s_%s_%s_%s.zip", repo, version, currentOS, runtime.GOARCH)

	downloadURL := ""
	for _, asset := range releases.Assets {
		if strings.HasSuffix(*asset.BrowserDownloadURL, zipName) {
			downloadURL = *asset.BrowserDownloadURL
			break
		}
	}

	if len(downloadURL) == 0 {
		return nil, fmt.Errorf("no release found for %s %s", s.Owner, s.Repo)
	}

	return &update.Asset{
		URL:           downloadURL,
		LatestVersion: version,
	}, nil
}
