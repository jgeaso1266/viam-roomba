//go:build linux

package viamroomba

import (
	"os"
	"syscall"
)

// flushRx discards any unread bytes from the serial receive buffer.
// This prevents stale bytes from corrupting subsequent sensor query responses.
func (c *roombaConn) flushRx() {
	f, ok := c.roomba.S.(*os.File)
	if !ok {
		return
	}
	const (
		tcflsh   = 0x540B
		tciflush = 0x00
	)
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(tcflsh), uintptr(tciflush))
}
