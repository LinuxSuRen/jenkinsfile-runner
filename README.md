# Jenkinsfile Runner

![Logo](logo.png)

## Introduction

Jenkinsfile Runner is an experiment to package Jenkins pipeline execution as a command line tool. The intend use cases include:

- Assist editing and testing Jenkinsfile locally
- Use Jenkins in Function-as-a-Service context
- Integration test shared libraries


## Build

Currently there's no released distribution, so you must first build the code by yourself:

```sh
mvn package
go install 
```

Assuming you have `$GOBIN` well set and declared in your `PATH`, you now have command line `jenkinsfile-runner` 
available to run from any directory containing a Jenkinsfile.   


## Usage

Jenkinsfile Runner is a command line tool you can invoke from any project directory containing a Jenkinsfile. It's a standalone
executable but require `java` in your PATH so it can run a Jenkins headless master.

Jenkinsfile Runner do:

- download latest Jenkins LTS
- install plugins as defined by a `plugins.txt` file in project directory. If non set it will install latest `workflow-aggregator`
- setup `.jenkinsfile-runner` directory within your project with a JENKINS_HOME to run your build
- run Jenkins master headless with a custom plugin installed to immediately run a single job based on local Jenkinsfile, then shutdown on completion.


## Implementation

Jenkinsfile Runner is *not* a re-implementation of the Pipeline execution library used by Jenkins. This would be a huge effort, and
anyway all the power of Pipeline comes from various Jenkins plugins to provide DSL keywords. Jenkinsfile-runner is
actually booting a (headless) Jenkins master, create a one-shot job to run once the Jenkinsfile script in your local 
directory, sending build log to `stdout`. 

Jenkinsfile Runner is a `go` program to setup a transient, headless Jenkins master. It automatically install the set of
plugins defined in a plugins.txt file stored aside your `Jenkinsfile`. It also will apply a `jenkins.yaml` 
[Configuration-as-Code](https://github.com/jenkinsci/configuration-as-code-plugin) file if present, which you can rely 
on to define the expected Jenkins configuration required for your pipeline script.

Within this transient jenkins master, a dedicated plugin is injected and will automatically create a job to run your
Jenkinsfile from local filesystem. This plugin also hacks Jenkins to redirect the build log to stdout, and shut down 
master when job is completed. 

