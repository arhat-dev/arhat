ARG ARCH=armv6

FROM arhatdev/base-go:debian-amd64 as builder
ARG ARCH=armv6

ENV CGO_ENABLED=1
RUN apk add --no-cache \
    python3-distutils=3.7.3-1 python3-lib2to3=3.7.3-1 python3=3.7.3-1

WORKDIR /app
COPY . /app
RUN make arhat-libpod.linux.${ARCH}

# since arhat-libpod uses cgo, alpine using musl-libc and debian glibc
# this container will not run, just for content delivery
FROM arhatdev/go:alpine-${ARCH}
ARG APP=arhat-libpod

ENTRYPOINT [ "/arhat-libpod" ]
