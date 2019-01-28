package common

import "errors"

var (
	ErrNoContainer = errors.New("No container")
	ErrNoImage     = errors.New("No image")
	ErrNoVolume    = errors.New("No volume")
	ErrNoNetwork   = errors.New("No network")
)
