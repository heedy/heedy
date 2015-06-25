package operator

import (
	"errors"
	"strings"
)

var (
	//ErrBadPath is thrown when a user device or stream cannot be extracted from the path
	ErrBadPath = errors.New("The given path is invalid.")
)

/** A path used for communication with Operators.

Paths are user[/device[/stream[/substream]]]
**/
type Path struct {
	User      string
	Device    string
	Stream    string
	Substream string

	// The length of the original path this was constructed from (number of /s)
	PathLen int
}

// Gets the path to the device this Path represents
func (p *Path) GetDevicePath() string {
	return p.User + "/" + p.Device
}

// Gets the path to the stream this Path represents
func (p *Path) GetStreamPath() string {
	return p.GetDevicePath() + "/" + p.Stream
}

// Gets the path to the substream this Path represents
func (p *Path) GetSubstreamPath() string {
	return p.GetStreamPath() + "/" + p.Substream
}

// Converts the given path into its components and returns it.
func CreatePath(path string) (Path, error) {
	split := strings.Split(path, "/")
	splitLen := len(split)

	var p Path
	p.PathLen = splitLen

	if 0 == splitLen {
		return p, ErrBadPath
	}

	if splitLen >= 4 {
		p.Substream = strings.Join(split[3:], "/")
	}

	if splitLen >= 3 {
		p.Stream = split[2]
	}

	if splitLen >= 2 {
		p.Device = split[1]
	}

	p.User = split[0]

	return p, nil
}

// SplitStreamPath splits a stream path to the given paths.
func SplitStreamPath(spath string) (username, devicepath, streampath, streamname, substream string, erro error) {
	path, err := CreatePath(spath)
	username = path.User
	devicepath = path.GetDevicePath()
	streampath = path.GetStreamPath()
	streamname = path.Stream
	substream = path.Substream

	if path.PathLen < 3 {
		err = ErrBadPath
	}

	return username, devicepath, streampath, streamname, substream, err
}

// SplitDevicePath splits the path into the device name and user name
func SplitDevicePath(dpath string) (username, devicename string, erro error) {
	path, err := CreatePath(dpath)
	username = path.User
	devicename = path.Device
	if path.PathLen != 2 {
		err = ErrBadPath
	}

	return username, devicename, err
}
