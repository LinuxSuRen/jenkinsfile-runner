FROM maven:3-jdk-8 as MavenBuilder

COPY /pom.xml /workspace/pom.xml
WORKDIR /workspace
RUN mvn dependency:go-offline
COPY /src /workspace/src
RUN mvn clean package


# ---

FROM golang:1.10 as GoBuilder

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
ENV WORKSPACE /go/src/github/ndeloof/jenkinsfile-runner
WORKDIR $WORKSPACE

COPY /Gopkg.* $WORKSPACE/
RUN dep ensure --vendor-only
COPY /*.go $WORKSPACE/
RUN go install

# ---

FROM openjdk:8

RUN wget -O go.tgz https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz ; \
    tar -C /usr/local -xzf go.tgz; \
    rm go.tgz;

ENV GOPATH /go
ENV PATH /usr/local/go/bin:$PATH

COPY --from=MavenBuilder /workspace/target/jenkinsfile-runner.hpi /usr/local/bin/jenkinsfile-runner.hpi
COPY --from=GoBuilder /go/bin/jenkinsfile-runner /usr/local/bin/jenkinsfile-runner

# /workspace should contain your project sources and Jenkinsfile
VOLUME /workspace
WORKDIR /workspace

# /var/jenkinsfile-runner will host transient jenkins master. Should be configured as a tmpfs volume
VOLUME /var/jenkinsfile-runner

# /var/jenkinsfile-runner-cache is used to cache downloaded jenkins.war and plugins. Should be shared|reused
VOLUME /var/jenkinsfile-runner-cache

CMD jenkinsfile-runner -workdir /var/jenkinsfile-runner -cache /var/jenkinsfile-runner-cache