// package goupdate provides tooling to auto-update binary releases
// from GitHub based on the user's current version and operating system.
package goupdate

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/c4milo/unpackit"
	"github.com/pkg/errors"
)

// Proxy is used to proxy a reader, for example
// using https://github.com/cheggaaa/pb to provide
// progress updates.
type Proxy func(int, io.ReadCloser) io.ReadCloser

// NopProxy does nothing.
var NopProxy = func(size int, r io.ReadCloser) io.ReadCloser {
	return r
}

// Manager is the update manager.
type Manager struct {
	Store          // Store for releases such as Github or a custom private store.
	Command string // Command is the executable's name.
}

// Release represents a project release.
type Release struct {
	Version     string    // Version is the release version.
	Notes       string    // Notes is the markdown release notes.
	URL         string    // URL is the notes url.
	PublishedAt time.Time // PublishedAt is the publish time.
	Assets      []*Asset  // Assets is the release assets.
}

// Asset represents a project release asset.
type Asset struct {
	Name      string // Name of the asset.
	Size      int    // Size of the asset.
	URL       string // URL of the asset.
	Downloads int    // Downloads count.
}

// InstallTo binary to the given dir.
func (m *Manager) InstallTo(path, dir string) error {
	log.Debugf("unpacking %q", path)

	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "opening tarball")
	}

	tempdir := filepath.Join(dir, "tmp")
	err = unpackit.Unpack(f, tempdir)
	if err != nil {
		f.Close()
		return errors.Wrap(err, "unpacking tarball")
	}
	defer os.RemoveAll(tempdir)

	if err := f.Close(); err != nil {
		return errors.Wrap(err, "closing tarball")
	}

	latestBinary := filepath.Join(tempdir, m.Command)

	if err := os.Chmod(latestBinary, 0755); err != nil {
		return errors.Wrap(err, "chmod")
	}

	currentBinary := filepath.Join(dir, m.Command)
	latestBinaryTmp := currentBinary + ".tmp"

	log.Debugf("copy %q to %q", latestBinary, latestBinaryTmp)
	if err := copyFile(latestBinaryTmp, latestBinary); err != nil {
		return errors.Wrap(err, "copying")
	}

	if runtime.GOOS == "windows" {
		old := currentBinary + ".old"
		log.Debugf("windows workaround renaming %q to %q", currentBinary, old)
		if err := os.Rename(currentBinary, old); err != nil {
			return errors.Wrap(err, "windows renaming")
		}
	}

	log.Debugf("renaming %q to %q", latestBinaryTmp, currentBinary)
	if err := os.Rename(latestBinaryTmp, currentBinary); err != nil {
		return errors.Wrap(err, "renaming")
	}

	return nil
}

// Install binary to replace the current version.
func (m *Manager) Install(path string) error {
	// bin, err := exec.LookPath(m.Command) // 获取当前程序的绝对路径
	// if err != nil {
	// 	return errors.Wrapf(err, "looking up path of %q", m.Command)
	// }

	// dir := filepath.Dir(bin)
	dir, err := getExecutablePath()
	if err != nil {
		return errors.Wrapf(err, "looking up path of %q", m.Command)
	}
	return m.InstallTo(path, dir)
}

// FindTarball returns a tarball matching os and arch, or nil.
func (r *Release) FindTarball(os, arch string) *Asset {
	s := fmt.Sprintf("%s_%s", os, arch)
	for _, a := range r.Assets {
		ext := filepath.Ext(a.Name)
		if strings.Contains(a.Name, s) && ext == ".gz" {
			return a
		}
	}

	return nil
}

// FindZip returns a zipfile matching os and arch, or nil.
func (r *Release) FindZip(os, arch string) *Asset {
	s := fmt.Sprintf("%s_%s", os, arch)
	for _, a := range r.Assets {
		ext := filepath.Ext(a.Name)
		if strings.Contains(a.Name, s) && ext == ".zip" {
			return a
		}
	}

	return nil
}

// Download the asset to a tmp directory and return its path.
func (a *Asset) Download() (string, error) {
	return a.DownloadProxy(NopProxy)
}

// DownloadProxy the asset to a tmp directory and return its path.
func (a *Asset) DownloadProxy(proxy Proxy) (string, error) {
	f, err := ioutil.TempFile(os.TempDir(), "update-")
	if err != nil {
		return "", errors.Wrap(err, "creating temp file")
	}

	log.Debugf("fetch %q", a.URL)
	res, err := http.Get(a.URL)
	if err != nil {
		return "", errors.Wrap(err, "fetching asset")
	}

	kind := res.Header.Get("Content-Type")
	size, _ := strconv.Atoi(res.Header.Get("Content-Length"))
	log.Debugf("response %s – %s (%d KiB)", res.Status, kind, size/1024)

	body := proxy(size, res.Body)

	if res.StatusCode >= 400 {
		body.Close()
		return "", errors.Wrap(err, res.Status)
	}

	log.Debugf("copy to %q", f.Name())
	if _, err := io.Copy(f, body); err != nil {
		body.Close()
		return "", errors.Wrap(err, "copying body")
	}

	if err := body.Close(); err != nil {
		return "", errors.Wrap(err, "closing body")
	}

	if err := f.Close(); err != nil {
		return "", errors.Wrap(err, "closing file")
	}

	log.Debugf("copied")
	return f.Name(), nil
}

// copyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func copyFile(dst, src string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}

	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}

	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func getExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return exePath, err
	}

	// 检查可执行文件是否存在
	_, err = os.Stat(exePath)
	if os.IsNotExist(err) {
		// 可执行文件不存在，采取适当的处理措施
		// 例如记录错误信息或提供备用路径
		return exePath, err
	}

	// 处理相对路径
	if !filepath.IsAbs(exePath) {
		exePath, err = filepath.Abs(exePath)
		if err != nil {
			// 处理转换为绝对路径时的错误
			return exePath, err
		}
	}

	// 获取可执行文件所在目录的绝对路径
	exeDir := filepath.Dir(exePath)

	return exeDir, err
}
