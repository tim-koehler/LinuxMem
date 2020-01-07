package linuxmem

import (
	"strconv"
	"syscall"
)

type MemoryHandler struct {
	pid int
}

func NewMemoryHandler(pid int) MemoryHandler {
	return MemoryHandler{pid: pid}
}

func (m MemoryHandler) ReadMemory(address int64, size int) ([]byte, error) {

	fd, err := attachAndSeekAddress(m.pid, address)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, size)
	_, err = syscall.Read(fd, buffer)
	if err != nil {
		closeAndDetach(fd, m.pid)
		return nil, err
	}

	err = closeAndDetach(fd, m.pid)

	return buffer, err
}

func (m MemoryHandler) WriteMemory(address int64, buffer []byte) error {

	fd, err := attachAndSeekAddress(m.pid, address)
	if err != nil {
		return err
	}

	_, err = syscall.Write(fd, buffer)
	if err != nil {
		closeAndDetach(fd, m.pid)
		return err
	}

	err = closeAndDetach(fd, m.pid)

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
