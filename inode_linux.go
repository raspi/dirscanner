package dirscanner

import (
	"syscall"
	"fmt"
	"os"
)

func getInode(path string) (uint64, error) {
	fi, err := os.Stat(path)

	if err != nil {
		return 0, fmt.Errorf(`couldn't stat'`)
	}
	stat, ok := fi.Sys().(*syscall.Stat_t)

	if !ok {
		return 0, fmt.Errorf(`couldn't stat'`)
	}

	return stat.Ino, nil

}
