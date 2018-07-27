# dirscanner
`dirscanner` is a recursive file lister which uses channels for go.

## Why?
When there's 1000000+ files in multiple directories crawling can take minutes. With a `dirscanner` channel you can start parsing files more quickly.

## Features

* You can provide a filter function to the scanner which validates what files will be sent for processing. For example: get only files that are between 1-10 MiB.

## Example usage:

```go
package main

import (
	"log"
	"time"
	"github.com/raspi/dirscanner"
	"runtime"
	"os"
	"sort"
)

// Example custom file validator
func validateFile(info os.FileInfo) bool {
	return info.Mode().IsRegular()
}

func main() {
	var err error

	workerCount := runtime.NumCPU()
	//workerCount := 1

	s := dirscanner.New()

	err = s.Init(workerCount, validateFile)
	if err != nil {
		panic(err)
	}
	defer s.Close()

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

scanloop:
	for {
		select {

		case <-s.Finished: // Finished getting file list
			log.Printf(`got all files`)
			break scanloop

		case e, ok := <-s.Errors: // Error happened, handle, discard or abort
			if ok {
				log.Printf(`got error: %v`, e)
				//s.Aborted <- true // Abort
			}


		case info, ok := <-s.Information: // Got information where worker is currently
			if ok {
				lastDir = info.Directory
			}


		case <-ticker.C: // Display some progress stats
			log.Printf(`%v Files scanned: %v Last file: %#v Dir: %#v`, time.Since(now).Truncate(time.Second), fileCount, lastFile, lastDir)

		case res, ok := <-s.Results:
			if ok {
				// Process file:
				lastFile = res.Path
				fileCount++
				sortedBySize[res.Size] = append(sortedBySize[res.Size], res.Path)
				//time.Sleep(time.Millisecond * 100)
			}
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
```
## Installation
To install this package, simply go get it:

    go get -u github.com/raspi/dirscanner

## Dependencies
There are no 3rd party package dependencies.

## Projects using this library
* https://github.com/raspi/duplikaatti
