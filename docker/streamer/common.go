package streamer

import "github.com/docker/docker/pkg/term"

type CommonStream struct {
	Fd         uintptr
	IsTerminal bool
	State      *term.State
}

func (s *CommonStream) RestoreTerminal() {
	if s.State != nil {
		term.RestoreTerminal(s.Fd, s.State)
	}
}
