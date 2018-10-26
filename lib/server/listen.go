package server

import (
	"bytes"
	"net"
	"os"
)

type Listener struct {
	Error error
	*net.UnixListener
	process ProcessFunc
	stopCh  chan struct{}
}

func NewListener(socket string, fn ProcessFunc) *Listener {
	os.Remove(socket)
	l, err := net.ListenUnix("unix", &net.UnixAddr{socket, "unix"})
	st := make(chan struct{}, 0)
	return &Listener{
		err, l, fn, st,
	}
}

type ProcessFunc func([]byte) []byte

func heard(c *net.UnixConn, l func([]byte) []byte) {
	var buf [1024]byte
	n, err := c.Read(buf[:])
	if err != nil {
		panic(err)
	}
	req := bytes.Trim(buf[:n], " ")
	resp := l(req)
	c.Write(resp)
	c.Close()
}

func (l *Listener) start() {
LISTEN:
	for {
		select {
		case <-l.stopCh:
			break LISTEN
		default:
			conn, err := l.AcceptUnix()
			if err != nil {
				panic(err)
			}
			go heard(conn, l.process)
		}
	}
}

func (l *Listener) stop() {
	l.Close()
	l.stopCh <- struct{}{}
}
