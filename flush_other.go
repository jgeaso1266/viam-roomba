//go:build !linux

package viamroomba

func (c *roombaConn) flushRx() {}
