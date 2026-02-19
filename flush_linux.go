//go:build linux

package viamroomba

import (
	"os"
	"syscall"
	"time"
	"unsafe"
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

// setReadTimeout configures the serial port so that read() returns after at
// most d (rounded to the nearest 100ms decisecond) instead of blocking forever.
// With VMIN=0, VTIME=N the kernel waits up to N*100ms for the first byte and
// returns 0 bytes (EOF in Go) if nothing arrives, releasing any mutex held by
// the caller.
func (c *roombaConn) setReadTimeout(d time.Duration) {
	f, ok := c.roomba.S.(*os.File)
	if !ok {
		return
	}

	// termios structure for Linux (x86-64 / arm64).
	// We only need the c_cc array (index 5 = VMIN, 6 = VTIME).
	type termios struct {
		Iflag  uint32
		Oflag  uint32
		Cflag  uint32
		Lflag  uint32
		Line   uint8
		Cc     [32]uint8
		Ispeed uint32
		Ospeed uint32
	}

	const (
		tcgets = 0x5401
		tcsets = 0x5402
		vmin   = 6
		vtime  = 5
	)

	var t termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(tcgets), uintptr(unsafe.Pointer(&t))); errno != 0 {
		return
	}

	deciseconds := uint8(d.Milliseconds() / 100)
	if deciseconds == 0 {
		deciseconds = 1
	}

	t.Cc[vmin] = 0
	t.Cc[vtime] = deciseconds

	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(tcsets), uintptr(unsafe.Pointer(&t)))
}
