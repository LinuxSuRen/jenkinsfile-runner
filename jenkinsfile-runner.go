package main

import (
	"fmt"
	"os"
	"os/exec"
	home "github.com/mitchellh/go-homedir"
	"flag"
	"path/filepath"
	"strings"
)

var jenkinsfile string
var version string
var cache string
var workdir string
var configfile string
var secretsfile string

type updateSitesFlag map[string]string

func (f *updateSitesFlag) Set(value string) error {
	i := strings.Index(value, "=")
	(*f)["@" + value[0:i]] = value[i+1:]
	return nil
}

func (f *updateSitesFlag) String() string {
	return "Update sites"
}

var updatesites updateSitesFlag = make(updateSitesFlag)

func main() {
	status := mainExitCode()
	if status != 0 {
		fmt.Println()
		fmt.Println("Something went wrong running your Jenkinsfile.")
		fmt.Println("If you think this is a bug, please report on https://github.com/ndeloof/jenkinsfile-runner/issues")
		fmt.Println("and if you miss some feature, please join https://gitter.im/jenkinsci/jenkinsfile-runner to discuss about it.")
	}
	os.Exit(status)
}

func mainExitCode() int {

	home, err := home.Dir()
	if err != nil {
		panic(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	updatesites["@default"] = "https://updates.jenkins.io"

	flag.StringVar(&jenkinsfile, "file", filepath.Join(wd, "Jenkinsfile"), "Jenkinsfile to run")
	flag.StringVar(&version, "version", "latest", "Jenkins version to use")
	flag.StringVar(&cache, "cache", filepath.Join(home, ".jenkinsfile-runner"), "Directory used as download cache")
	flag.StringVar(&configfile, "config", filepath.Join(wd, "jenkins.yaml"), "Configuration as Code file to setup jenkins master matching pipeline requirements")
	flag.StringVar(&secretsfile, "secrets", filepath.Join(wd, "secrets.gpg"), "GPG encrypted file containing sensitive data required to configure jenkins for your Pipeline")
	flag.Var(&updatesites, "site", "Update site to download plugins from. 'default=https://updates.jenkins.io/'")
	flag.StringVar(&workdir, "workdir", filepath.Join(wd, ".jenkinsfile-runner"), "Directory used to run headless jenkins master")

	flag.Parse()

	_, err = os.Stat(jenkinsfile)
	if os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr,"No such file %s\n", jenkinsfile)
		return -1
	}

	jenkinsfile, err = filepath.Abs(jenkinsfile)
	if err != nil {
		fmt.Fprintf(os.Stderr,"Can't find path to %s\n", jenkinsfile)
		return -1
	}
	mkdir(workdir)
	mkdir(cache)

	if version == "latest" {
		version = getLatestCoreVersion()
	}
    fmt.Printf("Running Pipeline on jenkins %s\n", version)


	war, err := getJenkinsWar(version)
	if err != nil {
		fmt.Fprintf(os.Stderr,"Failed to install Jenkins %s\n", version)
		return -1
	}

	mkdir(filepath.Join(workdir, "plugins"))
	err = installPlugins()
	if err != nil {
		fmt.Fprintf(os.Stderr,"Failed to install plugins: %s\n", err)
		return -1
	}
	err = InstallJenkinsfileRunner()
	if err != nil {
		fmt.Fprintf(os.Stderr,"Failed to install jenkinsfile-runner plugin: %s\n", err)
		return -1
	}

	secretsDir := filepath.Join(workdir, ".secrets")
	if _, err = os.Stat(secretsfile); err == nil {
		fmt.Printf("Using secrets from %s\n", secretsfile)
		secrets, err := decrypt(secretsfile)
		if err != nil {
			panic(err)
		}
		mkdir(secretsDir)
		propertiesToDockerSecretLayout(secrets, secretsDir)
		defer os.RemoveAll(secretsDir)
	}


	writeFile(filepath.Join(workdir, "logging.properties"), `
.level = INFO
handlers= java.util.logging.ConsoleHandler
java.util.logging.ConsoleHandler.level=WARNING
java.util.logging.ConsoleHandler.formatter=java.util.logging.SimpleFormatter`)

    fmt.Println("Starting Jenkins...")

	cmd := exec.Command("java",
		// "-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=5005",
		// disable setup wizard
		"-Djenkins.install.runSetupWizard=false",
		"-Djava.util.logging.config.file=.jenkinsfile-runner/logging.properties",
		"-jar", war, 
		// Disable http (so we can run in parallel without port collisions)
		"--httpPort=-1",
	)
	cmd.Env = append(os.Environ(),
		"JENKINS_HOME="+workdir,
		"JENKINSFILE="+jenkinsfile,
		"CASC_JENKINS_CONFIG="+configfile,
		"SECRETS="+secretsDir,
	)

	cmd.Stdout = os.Stdout	
	cmd.Stderr = os.Stderr	

	if err := cmd.Run(); err != nil {
		fmt.Printf("cmd.Start() failed with %s\n", err)
		return 1
	}
	return 0
}


func InstallJenkinsfileRunner() error {
	hpi := filepath.Join(workdir, "plugins", "jenkinsfile-runner.hpi")

	if _, err := os.Stat(hpi); err == nil {
		if err = os.Remove(hpi); err != nil {
			panic(err)
		}
	}	

	// TODO hpi file should be package within the jenkinsfile-runner binary as a "resource"
	// not sure about the preferred way to implement this in Go

	// We expect jenkinsfile-runner.hpi to be installed aside jenkinsfile-runner executable
	self, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return fmt.Errorf("Failed to retrieve jenkinsfile-runner installation path: %s", err)
	}

    if err := os.Link(filepath.Join(self, "jenkinsfile-runner.hpi"), hpi); err != nil {
		return err
    }
    return nil
}

