ARG ARCH=amd64

FROM arhatdev/builder-go:debian as builder
ENV CGO_ENABLED=1

FROM arhatdev/go:alpine-${ARCH}
ARG APP=arhat-libpod

ENTRYPOINT [ "/arhat-libpod" ]
