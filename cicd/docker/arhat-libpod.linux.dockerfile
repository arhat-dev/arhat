ARG ARCH=amd64

FROM arhatdev/base-go:debian-amd64 as builder
ARG ARCH=amd64

ENV CGO_ENABLED=1
RUN apt update ;\
    apt install -y python3-distutils=3.7.3-1 python3-lib2to3=3.7.3-1 python3=3.7.3-1

WORKDIR /app
COPY . /app
RUN make arhat-libpod.linux.${ARCH}

FROM arhatdev/go:debian-${ARCH}
ARG APP=arhat-libpod

ENTRYPOINT [ "/arhat-libpod" ]
