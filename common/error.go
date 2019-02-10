package common

import "errors"

var (
	// ErrNoContainer no container error.
	ErrNoContainer = errors.New("no container")
	// ErrNoImage no image error.
	ErrNoImage = errors.New("no image")
	// ErrNoVolume no volume error.
	ErrNoVolume = errors.New("no volume")
	// ErrNoNetwork no network error.
	ErrNoNetwork = errors.New("no network")
	// ErrDockerConnect cannot connect to docker engine error.
	ErrDockerConnect = errors.New("unable to connect to Docker")
	// ErrSmallTerminalWindowSize cannot run docui because of a small terminal window size
	ErrSmallTerminalWindowSize = errors.New("unable to run docui because of a small terminal window size")
)
