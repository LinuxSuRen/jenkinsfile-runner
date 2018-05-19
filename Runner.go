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
	"github.com/gnewton/jargo"
)

func main() {
    _, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	version := GetLatestCoreVersion()
    fmt.Printf("Running Pipeline on jenkins %s\n", version)

	_, err = os.Stat(".jenkinsfile-runner/plugins") 
	if os.IsNotExist(err) {
	    if err := os.MkdirAll(".jenkinsfile-runner/plugins", 0755); err != nil {
	        panic(err)
	    }
	}

	war, err := GetJenkinsWar(version)
	if err != nil {
		panic(err)
	}

	InstallPlugins()


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

func GetLatestCoreVersion() string {
	resp, err := http.Get("http://updates.jenkins.io/stable/latestCore.txt")
	if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
    	fmt.Printf("can't get latest code %s\n", resp.StatusCode)
    	os.Exit(1)
    }

    bodyBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    latest := string(bodyBytes)
    return latest;
}

func GetJenkinsWar(version string) (string, error) {
	war := fmt.Sprintf(".jenkinsfile-runner/jenkins-%s.war", version)
	_, err := os.Stat(war) 
	if os.IsNotExist(err) {

		fmt.Printf("Downloading jenkins %s...\n", version)

	    out, err := os.Create(war)
	    if err != nil {	    	
	        return war, err
	    }
		defer out.Close()

		url := fmt.Sprintf("http://updates.jenkins.io/download/war/%s/jenkins.war", version)
		resp, err := http.Get(url)
		if err != nil {	    	
	        panic(err)
	    }
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
	    	fmt.Printf("failed to download jenkins war. HTTP %s\n", resp.StatusCode)
	    	return war, err
	    }
	    size, err := strconv.Atoi(resp.Header.Get("Content-Length"))

	    ticker := time.NewTicker(time.Second * 2)
	    defer ticker.Stop()

	    go func() {
    		for _ = range ticker.C {
    			fi, err := os.Stat(war)
    			if err == nil {
    				downloaded := fi.Size()
	    			percent := 100 * float64(downloaded) / float64(size)
	    			fmt.Printf("%d (%.0f %%)\n", downloaded, percent)
    			}
    		}
    	}()	

		if _, err := io.Copy(out, resp.Body); err != nil {
	        return war, err
	    }
	}
	return war, nil;
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

func InstallPlugins() {

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
		dependsOn = append(dependsOn, InstallPlugin(shortname, version)...)
	}

	for len(dependsOn) > 0 {
		shortname := dependsOn[0]
		dependsOn = dependsOn[1:]
		if contains(installed, shortname) {
			continue
		} 
		dependsOn = append(dependsOn, InstallPlugin(shortname, "latest")...)
		installed = append(installed, shortname)
	}
}

func InstallPlugin(shortname string, version string) []string {

	hpi := fmt.Sprintf(".jenkinsfile-runner/plugins/%s-%s.hpi", shortname, version)
	_, err := os.Stat(hpi); 
	if os.IsNotExist(err) { 
	    out, err := os.Create(hpi)
	    if err != nil {	    	
	        panic(err)
	    }
		defer out.Close()

		url := fmt.Sprintf("http://updates.jenkins.io/download/plugins/%[1]s/%[2]s/%[1]s.hpi", shortname, version)
		fmt.Printf("Downloading %s:%s...\n", shortname, version)
		resp, err := http.Get(url)
		if err != nil {	    	
	        panic(err)
	    }
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
	    	fmt.Printf("failed to download %s plugin. HTTP %s\n", shortname, resp.StatusCode)
	    	panic(err)
	    }
	    size, err := strconv.Atoi(resp.Header.Get("Content-Length"))

	    ticker := time.NewTicker(time.Second * 2)
	    defer ticker.Stop()

	    go func() {
    		for _ = range ticker.C {
    			fi, err := os.Stat(hpi)
    			if err == nil {
    				downloaded := fi.Size()
	    			percent := 100 * float64(downloaded) / float64(size)
	    			fmt.Printf("%d (%.0f %%)\n", downloaded, percent)
    			}
    		}
    	}()	

		if _, err := io.Copy(out, resp.Body); err != nil {
	        panic(err)
	    }
    }

    manifest, err := jargo.GetManifest(hpi)
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

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}