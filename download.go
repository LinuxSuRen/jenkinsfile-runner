package main

import (
	"fmt"
	"os"
	"path/filepath"
	"net/http"
	"strconv"
	"time"
	"io"
	"errors"
)

// Download artifact specified by url to target file.
func download(url string, description string, target string) error {
	fmt.Printf("Downloading %s ...\n", description)
	download := target+".download"

	os.MkdirAll(filepath.Dir(target), 0755)

	// Check for a previous aborted download attempt
	if _, err := os.Stat(download); err == nil {
		if err = os.Remove(download); err != nil {
			return err
		}
	}

	out, err := os.Create(download)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		e := fmt.Sprintf("failed to download %s from %s: HTTP %d", description, url, resp.StatusCode)
		return errors.New(e)
	}
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))

	ticker := time.NewTicker(time.Second * 2)
	defer ticker.Stop()

	go func() {
		for _ = range ticker.C {
			fi, err := os.Stat(download)
			if err == nil {
				downloaded := fi.Size()
				percent := 100 * float64(downloaded) / float64(size)
				fmt.Printf("%10d (%.0f %%)\n", downloaded, percent)
			}
		}
	}()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	if _, err := os.Stat(target); err == nil {
		if err = os.Remove(target); err != nil {
			return err
		}
	}
	if err = os.Rename(download, target); err != nil {
		return err
	}
	return nil
}


