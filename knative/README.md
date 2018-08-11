# use jenkinsfile-runner as a knative Builder

As a transient, standalone executor for a Jenkinsfile, one can rely on jenkinsfile-runner to execute a Jenkins pipeline
as a Build step for knative.

[jenkins/jenkinsfile-runner](http://hub.docker.com/r/jenkins/jenkinsfile-runner) docker image is designed so it can be
used as a knative [Builder](https://github.com/knative/docs/blob/master/build/builder-contract.md)

## demo

if you haven't already, follow knative's documentation to install (at least) [knative build](https://github.com/knative/docs/blob/master/build/installing-build-component.md#adding-the-knative-build-component)

```sh
➜ kubectl apply -f https://storage.googleapis.com/knative-releases/build/latest/release.yaml
```

Apply sample build configuration on your kubernetes cluster:

```sh
➜ kubectl apply -f build-hello.yaml
build "hello" created
```

check created pod
```sh
➜ kubectl get pod
NAME          READY     STATUS     RESTARTS   AGE
hello-bbwdd   0/1       Init:2/3   0          17s
```

you can check created pod. This one will declare a container for scm checkout, credentials setup, and the configured build step:
```sh
➜ kubectl describe pod hello-bbwdd
(...)
  build-step-release:
    Container ID:   docker://79a23bcef73e2eaf9484862113edf3d59f2a94cec0a5f8138fc2c3616c25d3d8
    Image:          jenkins/jenkinsfile-runner:latest
    Image ID:       docker-pullable://jenkins/jenkinsfile-runner@sha256:f771a63ff1bd03d1e3cfdaa927798cb9fc1c5440df2c981a0e266118895d342c
    Port:           <none>
    State:          Terminated
      Reason:       Completed
      Exit Code:    0
      Started:      Sat, 11 Aug 2018 07:57:09 +0200
      Finished:     Sat, 11 Aug 2018 07:57:31 +0200
    Ready:          True
    Restart Count:  0
    Environment:
      HOME:  /builder/home
    Mounts:
      /builder/home from home (rw)
      /var/jenkinsfile-runner from jenkinsfile-runner (rw)
      /var/jenkinsfile-runner-cache from jenkinsfile-runner-cache (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-74p8t (ro)
      /workspace from workspace (rw)
(...)
```

you can grab build logs from pod's container `build-step-release` ("release" beeing the step name in `build-hello.yaml`)
```sh
➜ kubectl logs hello-bbwdd -c build-step-release
Downloading latestCore ...
Running Pipeline on jenkins 2.121.2
Starting Jenkins...
Running from: /var/jenkinsfile-runner-cache/war/jenkins-2.121.2.war
webroot: EnvVars.masterEnvVars.get("JENKINS_HOME")
Jenkins home directory: /var/jenkinsfile-runner found at: EnvVars.masterEnvVars.get("JENKINS_HOME")
Started
Running in Durability level: PERFORMANCE_OPTIMIZED
[Pipeline] node
Running on Jenkins in /var/jenkinsfile-runner/workspace/job
[Pipeline] {
[Pipeline] echo
Hello World
[Pipeline] }
[Pipeline] // node
[Pipeline] End of Pipeline
```

cleanup by deleting the Build definition
```
➜ kubectl delete Build hello
```