## goupdate

Goupdate 是用 Go 语言开发的工具，能够自动从 GitHub 和 Gitee 下载最新的发布版本，并更新本地程序。

本程序是基于 https://github.com/tj/go-update 进行的分支，进行了额外功能的增强。

## 更新 github
```go
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

```

## 更新 gitee

```go
package main

import (
	"github.com/zan8in/gologger"
	"github.com/zan8in/goupdate"
	"github.com/zan8in/goupdate/stores/gitee"
)

func main() {

	owner := "zanbin"
	repo := "afrog"
	version := "2.8.9"

	if result, err := gitee.Update(owner, repo, version); err != nil {
		gologger.Error().Msg(err.Error())
	} else {
		if result.Status == 2 {
			gologger.Info().Msgf("%s %s", repo, goupdate.LatestVersionTips)
		} else {
			gologger.Info().Msgf("Successfully updated to %s %s\n", repo, result.LatestVersion)
		}
	}

}

```