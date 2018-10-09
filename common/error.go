package common

import "errors"

var (
	NoContainer = errors.New("No container")
	NoImage     = errors.New("No image")
	NoVolume    = errors.New("No volume")
	NoNetwork   = errors.New("No network")
)
