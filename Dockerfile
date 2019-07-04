FROM golang:1.12-alpine3.9 AS builder
COPY . /go/app
RUN set -ex && \
    apk add --no-cache --no-progress libgit2-dev gcc git musl-dev && \
    cd /go/app && \
    go install

FROM alpine:3.9
COPY --from=builder /go/bin/concourse-git-resource /opt/bin/
RUN set -ex && \
    apk add --no-cache --no-progress libgit2 musl && \
    mkdir -p /opt/resource && \
    ln -s /opt/bin/concourse-git-resource /opt/resource/check && \
    ln -s /opt/bin/concourse-git-resource /opt/resource/in && \
    ln -s /opt/bin/concourse-git-resource /opt/resource/out
