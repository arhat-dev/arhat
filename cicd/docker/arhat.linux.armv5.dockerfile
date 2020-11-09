ARG ARCH=amd64

FROM ghcr.io/arhat-dev/builder-go:alpine as builder
FROM ghcr.io/arhat-dev/go:debian-${ARCH}
ARG APP=arhat

ENTRYPOINT [ "/arhat" ]
