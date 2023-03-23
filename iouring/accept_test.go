package iouring_test

import (
	"fmt"
	"net"
	"testing"
	"time"
	"unsafe"

	. "github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/3JoB/gain/iouring"
)

func TestAccept(t *testing.T) {
	ring, err := iouring.CreateRing()
	Nil(t, err)
	defer ring.Close()

	socketFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	Nil(t, err)
	err = unix.SetsockoptInt(socketFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
	Nil(t, err)
	err = unix.Bind(socketFd, &unix.SockaddrInet4{
		Port: testPort,
	})
	Nil(t, err)
	err = unix.SetNonblock(socketFd, false)
	Nil(t, err)
	err = unix.Listen(socketFd, 1)
	Nil(t, err)
	defer func() {
		err := unix.Close(socketFd)
		Nil(t, err)
	}()

	entry, err := ring.GetSQE()
	Nil(t, err)
	var clientLen = new(uint32)
	clientAddr := &unix.RawSockaddrAny{}
	*clientLen = unix.SizeofSockaddrAny
	clientAddrPointer := uintptr(unsafe.Pointer(clientAddr))
	clientLenPointer := uint64(uintptr(unsafe.Pointer(clientLen)))

	entry.PrepareAccept(int(uintptr(socketFd)), clientAddrPointer, clientLenPointer, 0)
	entry.UserData = uint64(socketFd)
	cqeNr, err := ring.Submit()
	Nil(t, err)
	Equal(t, uint(1), cqeNr)

	clientConnChan := make(chan net.Conn)
	go func() {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", testPort), time.Second)
		Nil(t, err)
		clientConnChan <- conn
	}()
	defer func() {
		conn := <-clientConnChan
		err = conn.(*net.TCPConn).SetLinger(0)
		Nil(t, err)
	}()

	cqes := make([]*iouring.CompletionQueueEvent, 128)
	Nil(t, err)
	for {
		n := ring.PeekBatchCQE(cqes)
		for i := 0; i < n; i++ {
			cqe := cqes[i]
			Equal(t, uint64(socketFd), cqe.UserData())
			Greater(t, cqe.Res(), int32(0))
			err = unix.Shutdown(int(cqe.Res()), unix.SHUT_RDWR)
			Nil(t, err)
		}
		if n > 0 {
			ring.CQAdvance(uint32(n))
			Nil(t, err)
			break
		}
	}
}
