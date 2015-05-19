package util

import (
	"errors"
	"strings"
)

var (
	//ErrBadPath is thrown when a user device or stream cannot be extracted from the path
	ErrBadPath = errors.New("The given path is invalid.")
)

//SplitStreamPath splits a stream path to the given paths.
func SplitStreamPath(spath string, err error) (username, devicepath, streampath, streamname, substream string, erro error) {
	if err != nil {
		return "", "", "", "", "", err
	}
	splitted := strings.Split(spath, "/")
	if len(splitted) < 3 {
		return "", "", "", "", "", ErrBadPath
	}
	username = splitted[0]
	devicepath = username + "/" + splitted[1]
	streamname = splitted[2]
	streampath = devicepath + "/" + streamname

	substream = ""
	if len(splitted) > 3 {
		substream = splitted[3]
		for i := 4; i < len(splitted); i++ {
			substream += "/" + splitted[i]
		}
	}

	return username, devicepath, streampath, streamname, substream, nil
}

//SplitDevicePath splits the path into the device name and user name
func SplitDevicePath(dpath string, err error) (username, devicename string, erro error) {
	if err != nil {
		return "", "", err
	}
	splitted := strings.Split(dpath, "/")
	if len(splitted) != 2 {
		return "", "", ErrBadPath
	}
	return splitted[0], splitted[1], nil
}
