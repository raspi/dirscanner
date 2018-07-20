package main

import (
	"log"
	"time"
	"github.com/raspi/dirscanner"
	"runtime"
	"os"
)

// Custom file validator
func validateFile(file os.FileInfo) bool {
	if !file.Mode().IsRegular() {
		return false
	}

	return true
}

func main() {
	var err error

	workerCount := uint64(runtime.NumCPU())
	//workerCount := uint64(1)

	s := dirscanner.New()

	err = s.Init(workerCount, validateFile)
	if err != nil {
		panic(err)
	}

	err = s.ScanDirectory(`/home/raspi/aur`)
	if err != nil {
		panic(err)
	}

	lastDir := ``
	lastFile := ``
	fileCount := uint64(0)

	// Ticker for stats
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	now := time.Now()

readloop:
	for {
		select {

		case <-s.Finished: // Finished getting file list
			log.Printf(`got all files`)
			break readloop

		case e := <-s.Errors: // Error happened, handle, discard or abort
			log.Printf(`got error: %v`, e)
			//s.Aborted <- true // Abort

		case info := <-s.Information: // Got information where worker is currently
			lastDir = info.Directory

		case <-ticker.C: // Display some progress stats
			log.Printf(`%v Files scanned: %v Last file: %#v Dir: %#v`, time.Since(now).Truncate(time.Second), fileCount, lastFile, lastDir)

		case res := <-s.Results:
			// Process file:
			lastFile = res.Path
			fileCount++
		}
	}

	log.Printf(`last: %v`, lastFile)
	log.Printf(`count: %v`, fileCount)

}
