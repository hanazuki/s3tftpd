# syntax = docker/dockerfile:experimental

FROM debian:bullseye as build-base
RUN apt-get update -qq && apt-get install -y --no-install-recommends devscripts wget
RUN gpg --no-default-keyring --keyring trustedkeys.gpg --fetch-keys https://github.com/hanazuki.gpg

FROM build-base as build-executile
WORKDIR /tmp/build
RUN dget https://github.com/hanazuki/executile/releases/download/v0.1.0/executile_0.1.0_source.changes
WORKDIR /tmp/build/executile-0.1.0
RUN apt-get build-dep -y .
RUN debuild -b -uc

FROM build-base as build-s3tftpd
WORKDIR /tmp/build/s3tftpd
COPY debian/control debian/
RUN apt-get build-dep -y .
COPY . .
RUN --mount=type=cache,target=/root/go/pkg/mod go mod vendor
RUN debuild -us -uc

FROM debian:bullseye
RUN --mount=type=bind,target=/tmp/build-executile,source=/tmp/build,from=build-executile \
    --mount=type=bind,target=/tmp/build-s3tftpd,source=/tmp/build,from=build-s3tftpd \
    apt-get update -qq && \
    apt-get install -y --no-install-recommends dumb-init /tmp/build-*/*.deb && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh

EXPOSE 69/udp
ENTRYPOINT ["dumb-init", "/docker-entrypoint.sh"]
