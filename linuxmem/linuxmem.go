package linuxmem

import (
	"strconv"
	"syscall"
)

type MemoryHandler struct {
	Pid       int
	BigEndian bool
}

func (m *MemoryHandler) ReadMemory(address int64, size int) ([]byte, error) {
	if err := attachToProcess(m.Pid); err != nil {
		return nil, err
	}

	fd, err := seekAddress(m.Pid, address)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, size)
	_, err = syscall.Read(fd, buffer)
	if err != nil {
		closeAndDetach(fd, m.Pid)
		return nil, err
	}

	if m.BigEndian {
		reverseBuffer(&buffer)
	}

	err = closeAndDetach(fd, m.Pid)
	return buffer, err
}

func (m *MemoryHandler) WriteMemory(address int64, buffer []byte) error {
	if err := attachToProcess(m.Pid); err != nil {
		return err
	}

	fd, err := seekAddress(m.Pid, address)
	if err != nil {
		return err
	}

	_, err = syscall.Write(fd, buffer)
	if err != nil {
		closeAndDetach(fd, m.Pid)
		return err
	}

	err = closeAndDetach(fd, m.Pid)
	return err
}

func attachToProcess(pid int) error {
	if err := syscall.PtraceAttach(pid); err != nil {
		return err
	}

	var wstat syscall.WaitStatus
	if _, err := syscall.Wait4(pid, &wstat, 0, nil); err != nil {
		syscall.PtraceDetach(pid)
		return err
	}
	return nil
}

func seekAddress(pid int, address int64) (int, error) {
	memFile := "/proc/" + strconv.Itoa(int(pid)) + "/mem"

	fd, err := syscall.Open(memFile, 2, 0)
	if err != nil {
		syscall.PtraceDetach(pid)
		return 0, err
	}

	_, err = syscall.Seek(fd, address, 0)
	if err != nil {
		closeAndDetach(fd, pid)
		return 0, err
	}

	return fd, err
}

func reverseBuffer(buffer *[]byte) {
	for i := len(*buffer)/2 - 1; i >= 0; i-- {
		opp := len(*buffer) - 1 - i
		(*buffer)[i], (*buffer)[opp] = (*buffer)[opp], (*buffer)[i]
	}
}

func closeAndDetach(fd int, pid int) error {
	if err := syscall.Close(fd); err != nil {
		syscall.PtraceDetach(pid)
		return err
	}
	if err := syscall.PtraceDetach(pid); err != nil {
		return err
	}
	return nil
}
