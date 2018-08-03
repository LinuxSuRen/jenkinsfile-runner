# Jenkinsfile Runner

[![Join the chat at https://gitter.im/jenkins/jenkinsfile-runner](https://img.shields.io/badge/%E2%8A%AA%20gitter%20-Join%20chat%20%E2%86%92-brightgreen.svg?style=flat)](https://gitter.im/jenkins/jenkinsfile-runner?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)



![Logo](logo.png)

## Introduction

Jenkinsfile Runner is an experiment to package Jenkins pipeline execution as a command line tool. The intend use cases include:

- Assist editing and testing Jenkinsfile locally
- Use Jenkins in Function-as-a-Service context
- Integration test shared libraries



## Usage

Jenkinsfile Runner is a command line tool you can invoke from any project directory containing a Jenkinsfile. It's a standalone
executable but require `java` in your PATH so it can run a Jenkins headless master.

Jenkinsfile Runner do:

- download latest Jenkins LTS
- setup a temporary JENKINS_HOME directory to run a headless jenkins for your Pipeline
- install plugins as defined by a `plugins.txt` file in project directory. If none set, it will install latest `workflow-aggregator`
- optionally create a `.secret` directory with secrets from a `secrets.gpg` GPG-encrypted file. Read mode on [sensitive data](#Sensitive data)  
- run Jenkins master headless with a custom plugin installed to immediately run a single job based on local Jenkinsfile, then shutdown on completion.

### Docker image

For your convenience you can use `jenkins/jenkinsfile-runner` docker image

```bash
$ docker run -it -v $(pwd):/workspace jenkins/jenkinsfile-runner
```` 

To avoid repeated download we suggest you define volumes for the download caches

```bash
$ docker run -it -v $(pwd):/workspace                         \
    -v jenkinsfile-runner-cache:/var/jenkinsfile-runner-cache \
    jenkins/jenkinsfile-runner
```` 

### Advanced docker image use and recommendations

We recommended to make the transient JENKINS_HOME a temporary volume so you don't consume all disk space with 
repeated builds:  

```bash
$ docker run -it -v $(pwd):/workspace                         \
    --tmpfs /var/jenkinsfile-runner                           \
    -v jenkinsfile-runner-cache:/var/jenkinsfile-runner-cache \
    jenkins/jenkinsfile-runner
```` 

Alternatively you might want to **reuse** this folder between runs, can be usefull to collect build output and
diagnose jenkinsfile-runner issues:

```bash
$ docker run -it -v $(pwd):/workspace                         \
    -v $(pwd)/jenkinsfile-runner:/var/jenkinsfile-runner                           \
    -v jenkinsfile-runner-cache:/var/jenkinsfile-runner-cache \
    jenkins/jenkinsfile-runner
```` 



## Build

Currently there's no released distribution, so you must first build the code by yourself:
You need Maven and Go SDK installed.
We also use [dep](https://github.com/golang/dep) for goland gependencies and vendor folder management.

You can either build with Dockerfile :
```sh
docker build -t jenkins/jenkinsfile-runner . 
```

or if you have adequate tools installed :

```sh
mvn install
dep ensure
go install 
```

Assuming you have `$GOBIN` well set and declared in your `PATH`, you now have command line `jenkinsfile-runner` 
available to run from any directory containing a Jenkinsfile.   

### Jenkins core version

You can choose the version of jenkins to run passing `-version` argument. Default value `latest` is an alias for
"_latest LTS release_" which is checked once a day. The requested jenkins.war is downloaded to [download cache](#cache) before
jenkins is started from local `.jenkisnfile-runner` JENKINS_HOME

### Plugins

You can include a `plugins.txt` file with plugins required to run your pipeline. Jenkinsfile-runner will 
download those plugins and dependencies into [download cache](#cache) and setup `.jenkisnfile-runner` JENKINS_HOME
accordingly.

`plugins.txt` file is a plain text format with a plugin per line, as
`<shortname>:<version>` 

If you use a custom update site to host your own plugins, you can suffix plugins with optional `@updatesiteId` and 
pass `-site` argument using `id=url` format to Jenkinsfile-runner.
 

Note: once [JENKINS-34002](https://issues.jenkins-ci.org/browse/JENKINS-34002) is implemented we will also
pick required dependencies from `Jenkinsfile`. 

### <a name="cache"></a>Download Cache

As Jenkinsfile-runner do download jenkins.war and plugins on-demand, it relies on a download cache.
Default location is `$HOME/.jenkinsfile-runner` but you can override using `-cache` option.


### Master Configuration

For non-trivial pipelines you'll need some way to configure Jenkins master, for sample to declare some
credentials referred by ID in your `Jenkinsfile`. Jenkinsfile-runner relies on 
[Configuration-as-Code](https://github.com/jenkinsci/jep/tree/master/jep/201) for this purpose. you only
need to provide a `jenkins.yaml` file aside your `Jenkinsfile` to have the transient jenkins master
configured accordingly. 

### Sensitive data

Your pipeline might require some sensitive data that you don't want to store in plain text in your SCM. 
For this purpose you can define them in a `secrets` file in java properties format, and encrypt it with
[GPG](https://www.gnupg.org/) allowing your teammates to decrypt this file: 

```bash
gpg --encrypt --recipient bob@acme.org --recipient alice@acme.org (...) secrets
``` 
This will produce an encrypted `secrets.gpg` file you can safely commit to SCM. **Never** commit the
initial secrets file (once encrypted, you can delete it). We highly recommend you add it to your
`.gitignore` to avoid mistakes.

We suggest you manage teammates gpg public keys in your SCM as a `keyring` to ensure everybody in the 
team knows each other GPG identity.

When such a `secrets.gpg` file exists aside your `Jenkinsfile` (or set by `-secrets` option) Jenkinsfile-runner
will create a `.secrets` directory in `.jenkisnfile-runner` JENKINS_HOME following Docker secrets layout, 
fully compatible with Configuration-as-Code so you can define your configuration with secret replacements:

```yaml
jenkins:
    somethingSecret: ${NEVER_STORED_IN_PLAIN_TEXT}
```  
To enforce security, this folder is deleted when Jenkinsfile-runner completes. 

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

