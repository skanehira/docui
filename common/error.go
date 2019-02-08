package common

import "errors"

var (
	// ErrNoContainer no container error.
	ErrNoContainer = errors.New("No container")
	// ErrNoImage no image error.
	ErrNoImage = errors.New("No image")
	// ErrNoVolume no volume error.
	ErrNoVolume = errors.New("No volume")
	// ErrNoNetwork no network error.
	ErrNoNetwork = errors.New("No network")
	// ErrDockerConnect cannot connect to docker engine error.
	ErrDockerConnect = errors.New("unable to connect to Docker")
)
