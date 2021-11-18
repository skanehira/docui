//go:build !windows
// +build !windows

package common

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

func getTermSize(fd uintptr) (int, int) {
	var sz struct {
		rows uint16
		cols uint16
	}

	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
		fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
	return int(sz.cols), int(sz.rows)
}

// IsTerminalWindowSizeThanZero check terminal window size
func IsTerminalWindowSizeThanZero() bool {
	out, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		Logger.Error(err)
		return false
	}

	defer out.Close()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGWINCH, syscall.SIGINT)

	for {
		// check terminal window size
		termw, termh := getTermSize(out.Fd())
		if termw > 0 && termh > 0 {
			return true
		}

		select {
		case signal := <-signalCh:
			switch signal {
			// when the terminal window size is changed
			case syscall.SIGWINCH:
				continue
			// use ctrl + c to cancel
			case syscall.SIGINT:
				return false
			}
		}
	}
}
