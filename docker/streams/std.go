package streams

import (
	"io"

	"github.com/docker/docker/pkg/term"
)

// Std have stdin, stdout, stderr
type Std struct {
	in  *In
	out *Out
	err io.Writer
}

// NewStd new std
func NewStd() *Std {
	stdin, stdout, stderr := term.StdStreams()

	return &Std{
		in:  NewIn(stdin),
		out: NewOut(stdout),
		err: stderr,
	}
}

// In return stdin
func (std *Std) In() *In {
	return std.in
}

// Out return stout
func (std *Std) Out() *Out {
	return std.out
}

// Err return stderr
func (std *Std) Err() io.Writer {
	return std.err
}
