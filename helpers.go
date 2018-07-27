package dirscanner

import (
	"os"
	"fmt"
	"path/filepath"
)

// Information about a file
type FileInformation struct {
	Path       string // Path to file
	Size       uint64 // File size
	Identifier uint64 // Identifier (inode)
	Mode       os.FileMode
}

func newFileInformation(path string, size uint64, id uint64, mode os.FileMode) FileInformation {
	return FileInformation{
		Path:       path,
		Size:       size,
		Identifier: id,
		Mode:       mode,
	}
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
func listFiles(dir string, fileValidatorFunc FileValidatorFunction) (files []FileInformation, directories []string, err error) {
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

			inode, err := getInode(fpath)
			if err != nil {
				continue
			}

			fi := newFileInformation(fpath, uint64(file.Size()), inode, file.Mode())

			if !fileValidatorFunc(fi) {
				// Not a valid file, continue
				continue
			}

			files = append(files, fi)

		}
	}

	return files, directories, nil
}
