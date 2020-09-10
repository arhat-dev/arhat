ARG ARCH=amd64

FROM arhatdev/builder-go:alpine as builder
ENV CGO_ENABLED=1

RUN apt-get update ;\
    apt-get install -y --no-install-recommends \
        gcc-arm-linux-gnueabi g++-arm-linux-gnueabi linux-libc-dev-armel-cross \
        gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf linux-libc-dev-armhf-cross \
        gcc-aarch64-linux-gnu g++-aarch64-linux-gnu linux-libc-dev-arm64-cross \
        gcc-powerpc64le-linux-gnu g++-powerpc64le-linux-gnu linux-libc-dev-ppc64el-cross \
        gcc-i686-linux-gnu g++-i686-linux-gnu linux-libc-dev-i386-cross \
        gcc-s390x-linux-gnu g++-s390x-linux-gnu linux-libc-dev-s390x-cross

FROM arhatdev/go:alpine-${ARCH}
ARG APP=arhat-libpod

ENTRYPOINT [ "/arhat-libpod" ]
