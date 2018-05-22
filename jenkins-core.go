package main

import (
	"path/filepath"
	"os"
	"io/ioutil"
	"fmt"
)

// Retrieve latest stable (LTS) jenkins core version
func getLatestCoreVersion() string {

	latest := filepath.Join(cache, "war", "latest.txt")
	info, err := os.Stat(latest)
	if os.IsNotExist(err) || needUpdate(info) {
		if err = download("http://updates.jenkins.io/stable/latestCore.txt", "latestCore", latest); err != nil {
			panic(err)
		}
	}

	bytes, err := ioutil.ReadFile(latest)
	if err != nil {
		panic(err)
	}

	return string(bytes);
}

// Retrieve jenkins.war artificat for specified version
func getJenkinsWar(version string) (string, error) {
	war := fmt.Sprintf("jenkins-%s.war", version)
	installed := filepath.Join(workdir, war)
	_, err := os.Stat(installed)
	if os.IsNotExist(err) {

		local := filepath.Join(cache, "war",  war)
		_, err = os.Stat(local)
		if os.IsNotExist(err) {
			url := fmt.Sprintf("http://updates.jenkins.io/download/war/%s/jenkins.war", version)
			if err = download(url, "Jenkins " + version, local); err != nil {
				return installed, err
			}
		}

		if _, err := os.Stat(installed); err == nil {
			if err = os.Remove(installed); err != nil {
				panic(err)
			}
		}
		if err = os.Symlink(local, installed); err != nil {
			return installed, err
		}
		return installed, nil

	}
	return installed, nil;
}
