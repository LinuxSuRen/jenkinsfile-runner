package main;

import (
	"fmt"
	"os"
	"os/exec"
	"net/http"
	"io"
	"io/ioutil"
	"strconv"
	"time"
	"bufio"
	"strings"
	"path/filepath"
	"github.com/gnewton/jargo"
	home "github.com/mitchellh/go-homedir"
)

func main() {

	home, err := home.Dir()
	if err != nil {
		panic(err)
	}

	cache := home + "/.jenkinsfile-runner"
	_, err = os.Stat(cache) 
	if os.IsNotExist(err) {
	    if err := os.MkdirAll(cache, 0755); err != nil {
	        panic(err)
	    }
	}

    _, err = os.Getwd()
	if err != nil {
		panic(err)
	}

	version := GetLatestCoreVersion(cache)
    fmt.Printf("Running Pipeline on jenkins %s\n", version)

	_, err = os.Stat(".jenkinsfile-runner/plugins") 
	if os.IsNotExist(err) {
	    if err := os.MkdirAll(".jenkinsfile-runner/plugins", 0755); err != nil {
	        panic(err)
	    }
	}

	war, err := GetJenkinsWar(version, cache)
	if err != nil {
		panic(err)
	}

	InstallPlugins(cache)


	InstallJenkinsfileRunner()


    if _, err = os.Stat(".jenkinsfile-runner/jenkins.log"); err == nil {
    	if err = os.Remove(".jenkinsfile-runner/jenkins.log"); err != nil {
			panic(err)
		}
    }

    if _, err = os.Stat(".jenkinsfile-runner/logging.properties"); err != nil {
    	
    	jul := []byte(`
.level = INFO
handlers= java.util.logging.ConsoleHandler
java.util.logging.ConsoleHandler.level=WARNING
java.util.logging.ConsoleHandler.formatter=java.util.logging.SimpleFormatter`)
		err := ioutil.WriteFile(".jenkinsfile-runner/logging.properties", jul, 0755)
			if err != nil {
			panic(err)
		}
    }



    fmt.Println("Starting Jenkins...")
	cmd := exec.Command("java", 
		"-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005",
		// disable setup wizard
		"-Djenkins.install.runSetupWizard=false",
		"-Djava.util.logging.config.file=.jenkinsfile-runner/logging.properties",
		"-jar", war, 
		// Disable http (so we can run in parallel without port collisions)
		"--httpPort=-1", 
		// redirect logs to a file, but then System.out is overriden to redirect to file
		// "--logfile=.jenkinsfile-runner/jenkins.log",		
		// "--debug=2", doesn't have any effet !
	)
	cmd.Env = append(os.Environ(), "JENKINS_HOME=.jenkinsfile-runner")

	cmd.Stdout = os.Stdout	
	cmd.Stderr = os.Stderr	

	if err := cmd.Run(); err != nil {
		fmt.Printf("cmd.Start() failed with %s\n", err)
		os.Exit(1)
	}
}

func GetLatestCoreVersion(cache string) string {

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

func GetJenkinsWar(version string, cache string) (string, error) {
	war := fmt.Sprintf("jenkins-%s.war", version)
	installed := filepath.Join(".jenkinsfile-runner", war)
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


func InstallJenkinsfileRunner() {
	if _, err := os.Stat(".jenkinsfile-runner/plugins/jenkinsfile-runner.hpi"); err == nil { 
		if err = os.Remove(".jenkinsfile-runner/plugins/jenkinsfile-runner.hpi"); err != nil {
			panic(err)
		}
	}	

    if err := os.Link("target/jenkinsfile-runner.hpi", ".jenkinsfile-runner/plugins/jenkinsfile-runner.hpi"); err != nil {
        panic(err)
    }
}

func InstallPlugins(cache string) {

	installed := []string{}	
	plugins := []string{}
	_, err := os.Stat("plugins.txt")
	if os.IsNotExist(err) { 
		plugins = append(plugins, "workflow-aggregator:latest")
	} else if os.IsNotExist(err) {
		// no default plugin selected, let's install workflow-aggregator	

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
		spec := strings.Split(s, ":")
		shortname := spec[0]
		installed = append(installed, shortname)
		version := spec[1]
		dependsOn = append(dependsOn, InstallPlugin(shortname, version, cache)...)
	}

	for len(dependsOn) > 0 {
		shortname := dependsOn[0]
		dependsOn = dependsOn[1:]
		if contains(installed, shortname) {
			continue
		} 
		dependsOn = append(dependsOn, InstallPlugin(shortname, "latest", cache)...)
		installed = append(installed, shortname)
	}
}

func InstallPlugin(shortname string, version string, cache string) []string {

	name := fmt.Sprintf("%s-%s.hpi", shortname, version)
	hpi := filepath.Join(".jenkinsfile-runner", "plugins", name)

	local := filepath.Join(cache, "plugins", shortname, name)	
	_, err := os.Stat(local); 
	if os.IsNotExist(err) { 
		url := fmt.Sprintf("http://updates.jenkins.io/download/plugins/%[1]s/%[2]s/%[1]s.hpi", shortname, version)
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
    	fmt.Printf("failed to download %s. HTTP %s\n", description, resp.StatusCode)
    	return err
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

func needUpdate(file os.FileInfo) bool {

	if file == nil {
		return true
	}

	// Check at least once a day
	return file.ModTime().Add(24 * 60 * 60 * 1000).Before(time.Now())
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}