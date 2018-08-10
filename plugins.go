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
func installPlugins() error {

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
	} else if err != nil {
		return fmt.Errorf("Failed to access plugins.txt: %s", err)
	} else {
        file, err := os.Open("plugins.txt")
        if err != nil {
            return fmt.Errorf("Failed to open plugins.txt: %s", err)
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            plugins = append(plugins, scanner.Text())
        }

        if err := scanner.Err(); err != nil {
            return fmt.Errorf("Failed to parse plugins.txt: %s", err)
        }
    }

	dependsOn := []string{}
	for _, s := range plugins {
		match := r.FindStringSubmatch(s)
		shortname := match[1]
		installed = append(installed, shortname)
		version := match[2]
		site := match[3]
		p, err := installPlugin(shortname, version, site)
		if err != nil {
			return fmt.Errorf("Failed to install %s:%s: %s", shortname, version, err)
		}
		dependsOn = append(dependsOn, p...)
	}

	for len(dependsOn) > 0 {
		shortname := dependsOn[0]
		dependsOn = dependsOn[1:]
		if contains(installed, shortname) {
			continue
		}
		p, err := installPlugin(shortname, "latest", "@default")
		if err != nil {
			return fmt.Errorf("Failed to install %s:%s: %s", shortname, version, err)
		}
		dependsOn = append(dependsOn, p...)
		installed = append(installed, shortname)
	}
	return nil
}

// Install a specific plugin : version using the cached artifact if present.
// return the list of required dependencies for installed plugin
func installPlugin(shortname string, version string, site string) ([]string, error) {

	if site == "" {
		site = "@default"
	}

	baseUrl := updatesites[site]
	if baseUrl == "" {
		return nil, fmt.Errorf("unknown update site %s, use -site option to set URL\n", site)
	}

	name := fmt.Sprintf("%s-%s.hpi", shortname, version)
	hpi := filepath.Join(workdir, "plugins", name)

	local := filepath.Join(cache, "plugins", shortname, name)
	_, err := os.Stat(local)
	if os.IsNotExist(err) {
		// TODO get plugin-versions metadata from update center to get actual download URL per version
		url := fmt.Sprintf("%[1]s/download/plugins/%[2]s/%[3]s/%[2]s.hpi", baseUrl, shortname, version)
		if err = download(url, shortname + ":" + version, local); err != nil {
			return nil, fmt.Errorf("Can't download plugin from %s: %s\n", url, err)
		}
	}

	if _, err := os.Stat(hpi); err == nil {
		if err = os.Remove(hpi); err != nil {
			return nil, fmt.Errorf("Failed to cleanup exising hpi file %s\n", err)
		}
	}
	if err = os.Symlink(local, hpi); err != nil {
		err = copy(local, hpi)
		if err != nil {
			return nil, fmt.Errorf("Failed to install hpi file %s\n", err)
		}
	}

	manifest, err := jargo.GetManifest(local)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse hpi file %s\n", err)
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
	return dependsOn, nil
}
