package dirscanner

import (
	"os"
	"fmt"
	"path/filepath"
)

// Information about a file
type FileInformation struct {
	Path string // Path to file
	Size uint64 // Size of file
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
func listFiles(dir string, fileValidatorFunc func(os.FileInfo) bool) (files []FileInformation, directories []string, err error) {
	directory, err := os.Open(dir)

	if err != nil {
		return []FileInformation{}, []string{}, err
	}

	fInfo, err := directory.Readdir(-1)
	directory.Close()
	if err != nil {
		return []FileInformation{}, []string{}, err
	}

	for _, file := range fInfo {
		fpath := filepath.Join(directory.Name(), file.Name())
		if file.IsDir() {
			directories = append(directories, fpath)
		} else {
			// Check file with given function
			if fileValidatorFunc(file) {
				files = append(files, FileInformation{
					Path: fpath,
					Size: uint64(file.Size()),
				})
			}
		}
	}

	return files, directories, nil
}
