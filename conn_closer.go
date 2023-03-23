package gain

import (
	"github.com/rs/zerolog"
	"golang.org/x/sys/unix"

	"github.com/3JoB/gain/iouring"
)

type connCloser struct {
	ring   *iouring.Ring
	logger zerolog.Logger
}

func (c *connCloser) addCloseRequest(fd int) (*iouring.SubmissionQueueEntry, error) {
	entry, err := c.ring.GetSQE()
	if err != nil {
		return nil, err
	}
	entry.PrepareClose(fd)
	entry.UserData = closeConnFlag | uint64(fd)
	return entry, nil
}

func (c *connCloser) addCloseConnRequest(conn *connection) (*iouring.SubmissionQueueEntry, error) {
	entry, err := c.addCloseRequest(conn.fd)
	if err != nil {
		return nil, err
	}
	conn.state = connClose
	return entry, nil
}

func (c *connCloser) syscallShutdownSocket(fileDescriptor int) error {
	return unix.Shutdown(fileDescriptor, unix.SHUT_RDWR)
}

func newConnCloser(ring *iouring.Ring, logger zerolog.Logger) *connCloser {
	return &connCloser{
		ring:   ring,
		logger: logger,
	}
}
