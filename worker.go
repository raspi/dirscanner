package dirscanner

import (
	"os"
	"fmt"
	"sync"
	"syscall"
	"time"
)

// How many directories to keep in queue
const DIRECTORY_QUEUE_SIZE = 65536

// Result of a worker
type workerResult struct {
	Path     string         // Path to file
	FileInfo os.FileInfo    // FileInfo
	Stat     syscall.Stat_t // stat'd information
}

// Send information of what worker is processing
type workerInfo struct {
	Directory string // Directory path
}

// File validator signature
type FileValidatorFunction func(path string, info os.FileInfo, stat syscall.Stat_t) bool

// Always use New() to get proper scanner
type DirectoryScanner struct {
	directoryScannerJobs chan string           // Jobs (scan directory X)
	Results              chan workerResult     // Results
	Finished             chan bool             // Scanner has finished?
	Aborted              chan bool             // Scanner has aborted?
	Information          chan workerInfo       // Information about scan progress
	Errors               chan error            // Errors that happened during scanning
	waitGroup            *sync.WaitGroup       // Waits jobs to be finished
	FileValidatorFunc    FileValidatorFunction // Function for file validation
	isInitialized        bool                  // Initializing function called?
	isFinished           bool                  // finished?
	isRecursive          bool                  // Scan recursively?
}

// Create new directory scanner
func New() DirectoryScanner {
	return DirectoryScanner{
		Results:              make(chan workerResult, 100),
		Finished:             make(chan bool, 1),
		Aborted:              make(chan bool, 1),
		Information:          make(chan workerInfo),
		waitGroup:            &sync.WaitGroup{},
		directoryScannerJobs: make(chan string, DIRECTORY_QUEUE_SIZE),
		Errors:               make(chan error),
		// Default validator:
		FileValidatorFunc: func(path string, info os.FileInfo, stat syscall.Stat_t) bool {
			// Accepts all files
			return true
		},
		isInitialized: false,
		isFinished:    false,
		isRecursive:   true,
	}
}

// Initialize workers
func (s *DirectoryScanner) Init(workerCount int, fileValidatorFunc FileValidatorFunction) (err error) {
	// Set file validator function which filters wanted files
	s.FileValidatorFunc = fileValidatorFunc

	if workerCount == 0 {
		return fmt.Errorf(`invalid amount of workers: %v`, workerCount)
	}

	// start N workers
	for i := 0; i < workerCount; i++ {
		go s.worker()
	}

	s.isInitialized = true

	return nil
}

// ScanDirectory scans given directory and send results (file paths) to a channel
func (s *DirectoryScanner) ScanDirectory(dir string) (err error) {
	if !s.isInitialized {
		return fmt.Errorf(`not initialized`)
	}

	if s.isFinished {
		return fmt.Errorf(`finished`)
	}

	s.isRecursive = true

	err = isDirectory(dir)

	if err != nil {
		return err
	}

	// Send directory to be scanned by a worker
	s.directoryScannerJobs <- dir

	// Add initial job
	s.waitGroup.Add(1)

	// When all jobs finished, shutdown the system.
	go func(sc *DirectoryScanner) {
		// Wait workers to be finished
		sc.waitGroup.Wait()

		for {
			// Wait queues to empty
			if len(sc.directoryScannerJobs) == 0 && len(sc.Results) == 0 {
				break
			}

			time.Sleep(time.Millisecond * 10)
		}

		sc.isFinished = true

		// Work is done
		sc.Finished <- true

	}(s)

	return nil
}

// Close channels
func (s *DirectoryScanner) Close() (err error) {
	// Close channels
	close(s.Finished)
	close(s.directoryScannerJobs)
	close(s.Results)
	close(s.Information)
	close(s.Aborted)
	close(s.Errors)

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
			s.waitGroup.Add(1)
			res := workerResult{
				Path:     file.Path,
				FileInfo: file.FileInfo,
				Stat:     file.Stat,
			}

			s.Results <- res

			s.waitGroup.Done()
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
