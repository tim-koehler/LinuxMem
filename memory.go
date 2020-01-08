package linuxmem

import (
	"strconv"
	"syscall"
)

type MemoryHandler struct {
	Pid       int
	BigEndian bool
}

func (m MemoryHandler) ReadMemory(address int64, size int) ([]byte, error) {

	fd, err := attachAndSeekAddress(m.Pid, address)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, size)
	_, err = syscall.Read(fd, buffer)
	if err != nil {
		closeAndDetach(fd, m.Pid)
		return nil, err
	}

	err = closeAndDetach(fd, m.Pid)

	if m.BigEndian {
		reverseBuffer(&buffer)
	}

	return buffer, err
}

func (m MemoryHandler) WriteMemory(address int64, buffer []byte) error {

	fd, err := attachAndSeekAddress(m.Pid, address)
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

func attachAndSeekAddress(pid int, address int64) (int, error) {

	memFile := "/proc/" + strconv.Itoa(int(pid)) + "/mem"

	err := syscall.PtraceAttach(pid)
	if err != nil {
		return 0, err
	}

	var wstat syscall.WaitStatus
	_, err = syscall.Wait4(pid, &wstat, 0, nil)
	if err != nil {
		syscall.PtraceDetach(pid)
		return 0, err
	}

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

	err := syscall.Close(fd)
	if err != nil {
		syscall.PtraceDetach(pid)
		return err
	}

	err = syscall.PtraceDetach(pid)
	if err != nil {
		return err
	}

	return nil
}
