package dirscanner

import (
	"os"
	"fmt"
	"errors"
	"sync"
)

const DIRECTORY_BUFFER = 65536

// Result of a worker
type workerResult struct {
	Path string
	Size uint64
}

// Send information of what worker is processing
type workerInfo struct {
	Directory string // Directory path
}

// Always use New() to get proper scanner
type DirectoryScanner struct {
	directoryScannerJobs chan string            // Jobs (scan directory X)
	Results              chan workerResult      // Results
	Finished             chan bool              // Scanner has finished?
	Aborted              chan bool              // Scanner has aborted?
	Information          chan workerInfo        // Information about scan progress
	Errors               chan error             // Errors that happened during scanning
	waitGroup            *sync.WaitGroup        // Waits jobs to be finished
	FileValidatorFunc    func(os.FileInfo) bool // Function for file validation
	isInitialized        bool                   // Initializing function called?
	isFinished           bool                   // finished?
	isRecursive          bool                   // Scan recursively?
}

// Create new directory scanner
func New() DirectoryScanner {
	return DirectoryScanner{
		Results:              make(chan workerResult, 1),
		Finished:             make(chan bool, 1),
		Aborted:              make(chan bool, 1),
		Information:          make(chan workerInfo),
		waitGroup:            &sync.WaitGroup{},
		directoryScannerJobs: make(chan string),
		Errors:               make(chan error),
		// Default validator:
		FileValidatorFunc: func(os.FileInfo) bool {
			// Accepts all files
			return true
		},
		isInitialized: false,
		isFinished:    false,
		isRecursive:   true,
	}
}

// Initialize workers
func (s *DirectoryScanner) Init(workerCount uint64, fileValidatorFunc func(info os.FileInfo) bool) (err error) {
	// Make buffer for scanner jobs
	s.directoryScannerJobs = make(chan string, DIRECTORY_BUFFER)

	// Set file validator function which filters wanted files
	s.FileValidatorFunc = fileValidatorFunc

	if workerCount == 0 {
		return errors.New(fmt.Sprintf(`invalid amount of workers: %v`, workerCount))
	}

	// start N workers
	for i := uint64(0); i < workerCount; i++ {
		go s.worker()
	}

	s.isInitialized = true

	return nil
}

func (s *DirectoryScanner) ScanDirectory(dir string) (err error) {
	if !s.isInitialized {
		return errors.New(fmt.Sprintf(`not initialized`))
	}

	if s.isFinished {
		return errors.New(fmt.Sprintf(`finished`))
	}

	s.isRecursive = true

	err = isDirectory(dir)

	if err != nil {
		return err
	}

	s.directoryScannerJobs <- dir

	// Add initial job
	s.waitGroup.Add(1)

	// When all jobs finished, shutdown the system.
	go func(sc *DirectoryScanner) {
		// Wait workers to be finished
		sc.waitGroup.Wait()

		sc.isFinished = true

		// Work is done
		sc.Finished <- true

		// Close channels
		close(sc.Finished)
		close(sc.directoryScannerJobs)
		close(sc.Results)
		close(sc.Information)
	}(s)

	return nil
}

// Worker which recursively iterates given directories
func (s *DirectoryScanner) worker() {
	for job := range s.directoryScannerJobs {
		// Send information what directory is being scanned
		info := workerInfo{
			Directory: job,
		}
		s.Information <- info

		files, dirs, err := listFiles(job, s.FileValidatorFunc)

		if err != nil {
			s.Errors <- err
		}

		// Got result(s) (files)
		for _, file := range files {
			res := workerResult{
				Path: file.Path,
				Size: file.Size,
			}

			s.Results <- res

		}

		if s.isRecursive {
			dirCount := len(dirs)

			if dirCount > 0 {
				// Add directory to job queue
				s.waitGroup.Add(dirCount)

				// Process directories with worker
				for _, dirname := range dirs {
					s.directoryScannerJobs <- dirname
				}
			}
		}

		// Directory scan job done
		s.waitGroup.Done()
	}

}
