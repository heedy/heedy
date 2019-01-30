package assets

import (
	"os"
	"syscall"
	"time"

	"github.com/gobuffalo/packr/v2"
	"github.com/gobuffalo/packr/v2/file"
	"github.com/spf13/afero"
)

// AferoPackr provides an afero.Fs interface to the packr Box
// This enables using Afero's overlay stuff
type AferoPackr struct {
	box *packr.Box
}

type BoxFile struct {
	file.File
}

func (b *BoxFile) WriteAt(p []byte, off int64) (int, error) {
	return 0, syscall.EPERM
}

func (b *BoxFile) ReadAt(p []byte, off int64) (int, error) {
	return 0, syscall.EPERM
}

func (b *BoxFile) Readdirnames(n int) ([]string, error) {
	dirs, err := b.File.Readdir(n)
	if err != nil {
		return nil, err
	}

	out := make([]string, len(dirs))
	for d := range dirs {
		out[d] = dirs[d].Name()
	}
	return out, nil
}

func (b *BoxFile) Sync() error {
	return nil
}

func (b *BoxFile) Truncate(size int64) error {
	return syscall.EPERM
}

func (b *BoxFile) WriteString(s string) (int, error) {
	return 0, syscall.EPERM
}

// NewAferoPackr takes a packr Box and creates an afero.Fs object
// compatible with Afero
func NewAferoPackr(box *packr.Box) afero.Fs {
	return &AferoPackr{box: box}
}

// Most of the code below is based off of afero's ReadOnly filesystem:
// https://github.com/spf13/afero/blob/master/readonlyfs.go

func (r *AferoPackr) ReadDir(name string) ([]os.FileInfo, error) {
	f, err := r.box.Open(name)
	if err != nil {
		return nil, err
	}
	return f.Readdir(0)
}

func (r *AferoPackr) Chtimes(n string, a, m time.Time) error {
	return syscall.EPERM
}

func (r *AferoPackr) Chmod(n string, m os.FileMode) error {
	return syscall.EPERM
}

func (r *AferoPackr) Name() string {
	return "AferoPackr"
}

func (r *AferoPackr) Stat(name string) (os.FileInfo, error) {
	f, err := r.box.Open(name)
	if err != nil {
		return nil, err
	}
	return f.Stat()
}

func (r *AferoPackr) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	fi, err := r.Stat(name)
	return fi, false, err
}

func (r *AferoPackr) Rename(o, n string) error {
	return syscall.EPERM
}

func (r *AferoPackr) RemoveAll(p string) error {
	return syscall.EPERM
}

func (r *AferoPackr) Remove(n string) error {
	return syscall.EPERM
}

func (r *AferoPackr) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if flag&(os.O_WRONLY|syscall.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		return nil, syscall.EPERM
	}

	return r.Open(name)
}

func (r *AferoPackr) Open(n string) (afero.File, error) {
	f, err := r.box.Resolve(n)
	return &BoxFile{f}, err
}

func (r *AferoPackr) Mkdir(n string, p os.FileMode) error {
	return syscall.EPERM
}

func (r *AferoPackr) MkdirAll(n string, p os.FileMode) error {
	return syscall.EPERM
}

func (r *AferoPackr) Create(n string) (afero.File, error) {
	return nil, syscall.EPERM
}
