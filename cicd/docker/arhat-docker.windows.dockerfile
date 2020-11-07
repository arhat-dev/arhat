ARG ARCH=amd64

FROM ghcr.io/arhat-dev/builder-go:alpine as builder
ENV CGO_ENABLED=0

# TODO: support multiarch build
FROM mcr.microsoft.com/windows/servercore:ltsc2019
ARG APP=arhat-docker

ENTRYPOINT [ "/arhat-docker" ]
