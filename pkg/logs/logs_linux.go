package logs

/* The purpose of this code is to log stderr (application panics) to a log file. */

import (
	"os"
	"syscall"
)

// nolint:gochecknoglobals
var stderr = os.Stderr.Fd()

func redirectStderr(file *os.File) {
	_ = syscall.Dup3(int(file.Fd()), int(stderr), 0)
}
