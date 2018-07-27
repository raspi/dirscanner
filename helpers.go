package dirscanner

import (
	"os"
	"fmt"
	"path/filepath"
	"syscall"
)

// Information about a file
type fileInformation struct {
	Path     string // Path to file
	FileInfo os.FileInfo
	Stat     syscall.Stat_t
}

// is directory and exists
func isDirectory(dir string) (err error) {
	fi, err := os.Stat(dir)

	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
	}

	if !fi.IsDir() {
		return fmt.Errorf(`not a directory: %v`, dir)
	}

	return nil
}

// List files and directories of given directory
// Filter accepted files with a function
func listFiles(dir string, fileValidatorFunc FileValidatorFunction) (files []fileInformation, directories []string, err error) {
	directory, err := os.Open(dir)

	if err != nil {
		return []fileInformation{}, []string{}, err
	}

	fInfo, err := directory.Readdir(-1)
	directory.Close()
	if err != nil {
		return []fileInformation{}, []string{}, err
	}

	for _, file := range fInfo {
		fpath := filepath.Join(directory.Name(), file.Name())
		if file.IsDir() {
			directories = append(directories, fpath)
		} else {

			stat, ok := file.Sys().(*syscall.Stat_t)

			if !ok {
				continue
			}

			if !fileValidatorFunc(fpath, file, *stat) {
				// Not a valid file, continue
				continue
			}

			files = append(files, fileInformation{
				Path:     fpath,
				FileInfo: file,
				Stat:     *stat,
			})

		}
	}

	return files, directories, nil
}
