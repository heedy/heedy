package assets

import (
	"os"
	"syscall"
	"time"

	"github.com/gobuffalo/packr/v2"
)

// AferoPackr provides an afero.Fs interface to the packr Box
// This enables using Afero's overlay stuff
type AferoPackr struct {
	box *packr.Box
}

// NewAferoPackr takes a packr Box and creates an afero.Fs object
// compatible with Afero
func NewAferoPackr(box *packr.Box) *AferoPackr {
	return &AferoPackr{box: box}
}

// Most of the code below is based off of afero's ReadOnly filesystem:
// https://github.com/spf13/afero/blob/master/readonlyfs.go

func (r *AferoPackr) ReadDir(name string) ([]os.FileInfo, error) {
	return ReadDir(r.source, name)
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
	return r.source.Stat(name)
}

func (r *AferoPackr) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	if lsf, ok := r.source.(Lstater); ok {
		return lsf.LstatIfPossible(name)
	}
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

func (r *AferoPackr) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	if flag&(os.O_WRONLY|syscall.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0 {
		return nil, syscall.EPERM
	}
	return r.source.OpenFile(name, flag, perm)
}

func (r *AferoPackr) Open(n string) (File, error) {
	return r.source.Open(n)
}

func (r *AferoPackr) Mkdir(n string, p os.FileMode) error {
	return syscall.EPERM
}

func (r *AferoPackr) MkdirAll(n string, p os.FileMode) error {
	return syscall.EPERM
}

func (r *AferoPackr) Create(n string) (File, error) {
	return nil, syscall.EPERM
}
