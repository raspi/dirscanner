package main

import (
	"log"
	"time"
	"github.com/raspi/dirscanner"
	"runtime"
	"os"
	"sort"
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

	err = s.ScanDirectory(`/home/raspi`)
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

	sortedBySize := map[uint64][]string{}

readloop:
	for {
		select {

		case <-s.Finished: // Finished getting file list
			log.Printf(`got all files`)
			break readloop

		case e, ok := <-s.Errors: // Error happened, handle, discard or abort
			if !ok {
				continue
			}

			log.Printf(`got error: %v`, e)
			//s.Aborted <- true // Abort

		case info, ok := <-s.Information: // Got information where worker is currently
			if !ok {
				continue
			}

			lastDir = info.Directory

		case <-ticker.C: // Display some progress stats
			log.Printf(`%v Files scanned: %v Last file: %#v Dir: %#v`, time.Since(now).Truncate(time.Second), fileCount, lastFile, lastDir)

		case res, ok := <-s.Results:
			if !ok {
				continue
			}

			// Process file:
			lastFile = res.Path
			fileCount++
			sortedBySize[res.Size] = append(sortedBySize[res.Size], res.Path)
			//time.Sleep(time.Millisecond * 100)
		}
	}

	// Sort
	var keys []uint64

	for k, _ := range sortedBySize {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	// Print in size order
	for _, k := range keys {
		log.Printf(`S:%v C:%v %v`, k, len(sortedBySize[k]), sortedBySize[k])
	}

	log.Printf(`last: %v`, lastFile)
	log.Printf(`count: %v`, fileCount)

}
