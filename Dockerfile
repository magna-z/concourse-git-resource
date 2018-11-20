FROM golang:1.11-stretch

RUN apt update && \
    apt install --yes \
    cmake \
    libcurl4-gnutls-dev \
    libhttp-parser-dev \
    libssh2-1-dev \
    libssl1.0-dev \
    libssh-dev \
    ssh-client \
    zlib1g-dev && \
    go get -d -v github.com/libgit2/git2go && \
    cd $GOPATH/src/github.com/libgit2/git2go && \
    git checkout origin/v27 && \
    git submodule update --init && \
    make build-libgit2 && \
    mkdir -p /opt/resource/

COPY . /go/src/github.com/devinotelecom/concourse-git-resource

RUN go get -d -v github.com/sirupsen/logrus && \
    go build --tags "static" -ldflags "-X main.version=1.0.0" -o /opt/resource/check $GOPATH/src/github.com/devinotelecom/concourse-git-resource/check && \
    go build --tags "static" -ldflags "-X main.version=1.0.0" -o /opt/resource/in $GOPATH/src/github.com/devinotelecom/concourse-git-resource/in
