//go:build !windows

package pidfile

import "syscall"

// signalZero is the signal number for liveness checks.
// On Unix, sending signal 0 to a process checks if it
// exists without actually delivering a signal.
const signalZero = syscall.Signal(0)
