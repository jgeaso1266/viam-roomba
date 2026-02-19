package viamroomba

import (
	"fmt"
	"sync"
	"time"

	"github.com/parabolala/go-roomba"
)

type roombaConn struct {
	roomba *roomba.Roomba
	mu     sync.Mutex
	refs   int
}

var (
	globalMu    sync.Mutex
	connections = map[string]*roombaConn{}
)

func acquireConn(serialPort string) (*roombaConn, error) {
	globalMu.Lock()
	defer globalMu.Unlock()
	if conn, ok := connections[serialPort]; ok {
		conn.refs++
		return conn, nil
	}
	r, err := roomba.MakeRoomba(serialPort)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial connection on %s: %w", serialPort, err)
	}
	// Send START command (opcode 128) to enable the Open Interface before any queries or commands.
	if err := r.Passive(); err != nil {
		return nil, fmt.Errorf("failed to start OI on %s: %w", serialPort, err)
	}
	conn := &roombaConn{roomba: r, refs: 1}
	conn.setReadTimeout(2 * time.Second)
	connections[serialPort] = conn
	return conn, nil
}

func releaseConn(serialPort string) {
	globalMu.Lock()
	defer globalMu.Unlock()
	conn, ok := connections[serialPort]
	if !ok {
		return
	}
	conn.refs--
	if conn.refs <= 0 {
		delete(connections, serialPort)
	}
}
