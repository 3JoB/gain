package gain

import (
	"errors"
	"time"

	"golang.org/x/sys/unix"

	"github.com/3JoB/gain/iouring"
)

type singleSubmitter struct {
	ring            *iouring.Ring
	timeoutTimeSpec unix.Timespec
}

func (s *singleSubmitter) submit() error {
	_, err := s.ring.SubmitAndWaitTimeout(1, &s.timeoutTimeSpec)
	if errors.Is(err, iouring.ErrAgain) || errors.Is(err, iouring.ErrInterrupredSyscall) ||
		errors.Is(err, iouring.ErrTimerExpired) {
		return errSkippable
	}
	return err
}

func (s *singleSubmitter) advance(n uint32) {
	s.ring.CQAdvance(n)
}

func newSingleSubmitter(ring *iouring.Ring) *singleSubmitter {
	return &singleSubmitter{
		ring:            ring,
		timeoutTimeSpec: unix.NsecToTimespec((time.Millisecond).Nanoseconds()),
	}
}
