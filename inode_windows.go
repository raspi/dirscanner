package dirscanner

import (
	"syscall"
	"os"
)

func getInode(path string) (uint64, error) {
	pathptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, os.NewSyscallError("UTF16PtrFromString", err)
	}
	h, e := syscall.CreateFile(pathptr, 0, 0, nil, syscall.OPEN_EXISTING, 0, 0)

	if e != nil {
		return 0, os.NewSyscallError("CreateFile", e)
	}

	var fi syscall.ByHandleFileInformation
	if e = syscall.GetFileInformationByHandle(h, &fi); e != nil {
		syscall.CloseHandle(h)
		return 0, os.NewSyscallError("GetFileInformationByHandle", e)
	}

	return uint64(fi.FileIndexHigh)<<32 | uint64(fi.FileIndexLow), nil
}
