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
)

func main() {
	fmt.Println("Jenkinsfile Runner")

	version := GetLatestCoreVersion()
    fmt.Printf("Running Pipeline on jenkins %s\n", version)

	_, err := os.Stat(".jenkinsfile-runner") 
	if os.IsNotExist(err) {
	    if err := os.Mkdir(".jenkinsfile-runner", 0755); err != nil {
	        panic(err)
	    }
	}

	war := GetJenkinsWar(version)


	cmd := exec.Command("java", "-jar", war)

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

func GetJenkinsWar(version string) string {
	war := fmt.Sprintf(".jenkinsfile-runner/jenkins-%s.war", version)
	_, err := os.Stat(war) 
	if os.IsNotExist(err) {

		fmt.Printf("Downloading jenkins %s...\n", version)

	    out, err := os.Create(war)
	    if err != nil {
	        panic(err)
	    }
		defer out.Close()

		url := fmt.Sprintf("http://updates.jenkins.io/download/war/%s/jenkins.war", version)
		resp, err := http.Get(url)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
	    	fmt.Printf("failed to download jenkins war. HTTP %s\n", resp.StatusCode)
	    	os.Exit(1)
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
	        panic(err)
	    }
	}
	return war;
}
