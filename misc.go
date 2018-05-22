package main

import (
	"os"
	"time"
	"io/ioutil"
)

func mkdir(path string) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			panic(err)
		}
	}
}

func writeFile(path string, content string) {
	err := ioutil.WriteFile(path, []byte(content), 0755)
	if err != nil {
		panic(err)
	}
}

// Check if metadata within target file are up-to-date or require refresh
func needUpdate(file os.FileInfo) bool {

	if file == nil {
		return true
	}

	// Check at least once a day
	return file.ModTime().Add(24 * 60 * 60 * 1000).Before(time.Now())
}

// Check a slice do container specified element.
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
