package gain

import (
	"golang.org/x/sys/unix"
)

type buffer struct {
	data   []byte
	iovecs []unix.Iovec
}

func createBuffer(bufferSize uint) *buffer {
	bufferMem := make([]byte, bufferSize)
	iovec := unix.Iovec{
		Base: &bufferMem[0],
		Len:  uint64(bufferSize),
	}
	buff := &buffer{
		data:   bufferMem,
		iovecs: []unix.Iovec{iovec},
	}
	return buff
}

func createBuffers(bufferSize uint, maxConn uint) []*buffer {
	memSize := maxConn * bufferSize
	bufferMem := make([]byte, memSize)
	buffs := make([]*buffer, maxConn)
	for index := range buffs {
		startIndex := index * int(bufferSize)
		buff := bufferMem[startIndex : startIndex+int(bufferSize)]
		iovec := unix.Iovec{
			Base: &buff[0],
			Len:  uint64(bufferSize),
		}
		buffs[index] = &buffer{
			data:   buff,
			iovecs: []unix.Iovec{iovec},
		}
	}
	return buffs
}
