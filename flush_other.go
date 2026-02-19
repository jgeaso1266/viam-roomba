//go:build !linux

package viamroomba

import "time"

func (c *roombaConn) flushRx() {}

func (c *roombaConn) setReadTimeout(_ time.Duration) {}
