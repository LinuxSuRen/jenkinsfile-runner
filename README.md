# Jenkinsfile Runner

![Logo](logo.png)

## Introduction

Jenkinsfile Runner is an experiment to package Jenkins pipeline execution as a command line tool. The intend use cases include:

- Assist editing and testing Jenkinsfile locally
- Use Jenkins in Function-as-a-Service context
- Integration test shared libraries


## Build

Currently there's no released distribution, so you must first build the code by yourself:
You need Maven and Go SDK installed

```sh
make build 
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
- create a `.secret` directory with secrets from a `secrets.gpg` GPG-encrypted file.  
- run Jenkins master headless with a custom plugin installed to immediately run a single job based on local Jenkinsfile, then shutdown on completion.


### Jenkins core version

You can choose the version of jenkins to run passing `-version` argument. Default value `latest` is an alias for
"_latest LTS release_" which is checked once a day. The requested jenkins.war is downloaded to download cache before
jenkins is started from local `.jenkisnfile-runner` JENKINS_HOME

### Plugins

You can include a `plugins.txt` file with plugins required to run your pipeline. Jenkinsfile-runner will 
download those plugins and dependencies into dowload cache and setup `.jenkisnfile-runner` JENKINS_HOME
accordingly.

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

When such a `secrets.gpg` file exists aside your `Jenkinsfile` (or set by `-secrets option) Jenkinsfile-runner
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

