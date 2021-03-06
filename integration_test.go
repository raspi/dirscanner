package dirscanner

import (
	"testing"
	"io/ioutil"
	"os"
	"path/filepath"
)

func unique(s []string) (diff []string) {
	encountered := map[string]int{}

	for _, v := range s {
		encountered[v]++
	}

	for _, v := range s {
		if encountered[v] == 1 {
			diff = append(diff, v)
		}
	}

	return

}

func TestIntegration(t *testing.T) {
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tempDir, err := ioutil.TempDir(current, `test`)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	t.Logf(`using temp dir: %#v`, tempDir)

	createdTempRootDirectories := []string{
		``,
		`first`,
		`second`,
		`INTEGRATION TEST!!`,
	}

	createdTempDirectories := []string{
		``,
		`test1`,
		`test2`,
		`hello world`,
		`.hidden`,
	}

	createdTempFiles := []string{
		`a.txt`,
		`hello world.txt`,
		`foo.txt`,
		`test.txt`,
		`.hideme`,
		`__test`,
	}

	for _, tempRoot := range createdTempRootDirectories {
		for _, createDir := range createdTempDirectories {
			for _, createFile := range createdTempFiles {
				fpath := filepath.Join(tempDir, tempRoot, createDir, createFile)
				dir := filepath.Dir(fpath)

				err = os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					t.Fatal(err)
				}

				f, err := os.Create(fpath)
				if err != nil {
					t.Fatal(err)
				}

				for i := 0; i < 100; i++ {
					f.Write([]byte(`This is a test file! `))
				}

				f.Close()
			}
		}
	}

	var expectedTotalFilesize = uint64(0)
	var expectedFiles []string

	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			expectedFiles = append(expectedFiles, path)
			expectedTotalFilesize += uint64(info.Size())
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Test scanner
	scanner := New()
	err = scanner.Init(1, func(info os.FileInfo) bool {
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	err = scanner.ScanDirectory(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	var actualTotalFilesize = uint64(0)
	var actualFiles []string

readloop:
	for {
		select {

		case <-scanner.Finished:
			break readloop

		case e := <-scanner.Errors:
			t.Fatal(e)

		case _ = <-scanner.Information:

		case res, ok := <-scanner.Results:
			if !ok {
				continue
			}

			actualFiles = append(actualFiles, res.Path)
			actualTotalFilesize += res.Size
		}
	}

	expectedFileCount := len(expectedFiles)
	actualFileCount := len(actualFiles)

	if actualFileCount != expectedFileCount {
		diff := unique(append(actualFiles, expectedFiles...))
		t.Logf(`expected %v files, got %v`, expectedFileCount, actualFileCount)
		t.Logf(`Diff: %v`, diff)
		t.Fail()
	}

	if actualTotalFilesize != expectedTotalFilesize {
		t.Logf(`expected total size: %v, got %v`, expectedTotalFilesize, actualTotalFilesize)
		t.Fail()
	}


}
