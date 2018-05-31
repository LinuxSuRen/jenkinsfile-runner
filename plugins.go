package main

import (
	"github.com/gnewton/jargo"
	"fmt"
	"path/filepath"
	"os"
	"strings"
	"bufio"
	"regexp"
)

// Install all plugins required by the target Jenkinsfile.
// If no metadata (plugins.txt) has been specified it will install workflow-aggregator-plugin
func installPlugins() {

	installed := []string{}
	plugins := []string{}
	r := regexp.MustCompile(`^(.*):([^@]*)(@.*)?$`)

	/* prepare JENKINS-34002
		b, err := ioutil.Readfile("Jenkinsfile")
		if err != nil {
			panic(err)
		}
		s := string(b)

		i := strings.Index(txt, "requirePlugins")
		if i >= 0 {
	...
		}
	*/

	_, err := os.Stat("plugins.txt")
	if os.IsNotExist(err) {
		// no plugin specified, let's install workflow-aggregator
		plugins = append(plugins, "workflow-aggregator:latest")
	} else if err == nil {

		file, err := os.Open("plugins.txt")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			plugins = append(plugins, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}
	} else {
		panic(err)
	}

	dependsOn := []string{}
	for _, s := range plugins {
		match := r.FindStringSubmatch(s)
		shortname := match[1]
		installed = append(installed, shortname)
		version := match[2]
		site := match[3]
		dependsOn = append(dependsOn, installPlugin(shortname, version, site)...)
	}

	for len(dependsOn) > 0 {
		shortname := dependsOn[0]
		dependsOn = dependsOn[1:]
		if contains(installed, shortname) {
			continue
		}
		dependsOn = append(dependsOn, installPlugin(shortname, "latest", "default")...)
		installed = append(installed, shortname)
	}
}

// Install a specific plugin : version using the cached artifact if present.
// return the list of required dependencies for installed plugin
func installPlugin(shortname string, version string, site string) []string {

	if site == "" {
		site = "@default"
	}

	baseUrl := updatesites[site]
	if baseUrl == "" {
		fmt.Printf("unknown update site %s, use -site option to set URL\n", site)
		os.Exit(66)
	}

	name := fmt.Sprintf("%s-%s.hpi", shortname, version)
	hpi := filepath.Join(workdir, "plugins", name)

	local := filepath.Join(cache, "plugins", shortname, name)
	_, err := os.Stat(local);
	if os.IsNotExist(err) {
		url := fmt.Sprintf("%[1]s/download/plugins/%[2]s/%[3]s/%[2]s.hpi", baseUrl, shortname, version)
		if err = download(url, shortname + ":" + version, local); err != nil {
			panic(err)
		}
	}

	if _, err := os.Stat(hpi); err == nil {
		if err = os.Remove(hpi); err != nil {
			panic(err)
		}
	}
	if err = os.Symlink(local, hpi); err != nil {
		panic(err)
	}

	manifest, err := jargo.GetManifest(local)
	if err != nil {
		panic(err)
	}

	dependencies := (*manifest)["Plugin-Dependencies"]
	dependsOn := []string{}
	for _, d := range strings.Split(dependencies, ",") {
		if strings.Contains(d, "optional=true") {
			continue
		}
		d = strings.Replace(d, " ", "", -1)
		if len(d) == 0 {
			continue
		}
		dependsOn = append(dependsOn, strings.Split(d, ":")[0])
	}
	return dependsOn
}
