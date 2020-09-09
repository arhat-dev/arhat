ARG ARCH=amd64

FROM arhatdev/builder-go:alpine as builder
# TODO: support multiarch build
FROM mcr.microsoft.com/windows/servercore:ltsc2019
ARG APP=template-application-go

ENTRYPOINT [ "/template-application-go" ]
