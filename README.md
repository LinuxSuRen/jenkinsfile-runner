# Jenkinsfile Runner

![Logo](logo.png)

## Introduction

If you adopted `Jenkinsfile` do define your CI/CD workflow you probably have a git history which looks like this :

```bash
843ee85 - (HEAD -> master) hum 
d405868 - eventually got Jenkinsfile fixed 
70f99be - fix 
b101522 - oups 
1c21e7d - fix Jenkinsfile (again) 
5faca57 - fix Jenkinsfile 
a33d5e6 - Update Jenkinsfile to introduce xxx 
```

You indeed can't check your Jenkinsfile is correct until it has been pushed to SCM and built by Jenkins.
Generaly speaking this is annoying one can't run a Jenkinsfile _but_ within a full featured Jenkins master.

Jenkinsfile-runner intent to fix this by providing a command line tool to run a Jenkinsfile.

This is *not* a re-implementation of the Pipeline execution library used by Jenkins. This would be a huge effort, and
anyway all the power of Pipeline comes from various Jenkins plugins to provide DSL keywords. Jenkinsfile-runner is
actually booting a (headless) Jenkins master, create a one-shot job to run once the Jenkinsfile script in your local 
directory, sending build log to `stdout`. 

As a result of this machinery you get your Jenkinsfile running from CLI, within a _real_ jenkins with the exact set
of plugins and configuration your project has been designed for (more of this later). And you can then push to `master`
with confidence (or don't need to `push --force` or squash-merge to pretend you did it right with a single commit).

## Implementation

Jenkinsfile Runner is a `go` program to setup a transient, headless Jenkins master. It automatically install the set of
plugins defined in a plugins.txt file stored aside your `Jenkinsfile`. It also will apply a `jenkins.yaml` 
[Configuration-as-Code](https://github.com/jenkinsci/configuration-as-code-plugin) file if present, which you can rely 
on to define the expected Jenkins configuration required for your pipeline script.

Within this transient jenkins master, a dedicated plugin is injected and will automatically create a job to run your
Jenkinsfile from local filesystem. This plugin also hacks Jenkins to redirect the build log to stdout, and shut down 
master when job is completed. 
