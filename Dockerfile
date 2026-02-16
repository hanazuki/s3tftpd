# syntax = docker/dockerfile:1


FROM debian:trixie AS build-base

RUN apt-get update -qq && apt-get install -y --no-install-recommends devscripts wget dirmngr
RUN gpg --no-default-keyring --keyring trustedkeys.gpg --fetch-keys https://github.com/hanazuki.gpg

FROM build-base AS build-executile
WORKDIR /tmp/build
RUN dget https://github.com/hanazuki/executile/releases/download/v0.1.2/executile_0.1.2_source.changes
WORKDIR /tmp/build/executile-0.1.2
RUN apt-get build-dep -y .
RUN debuild -b -uc

FROM build-base AS build-s3tftpd
WORKDIR /tmp/build/s3tftpd
COPY debian/control debian/
RUN apt-get build-dep -y .
COPY . .
RUN --mount=type=cache,target=/root/go/pkg/mod go mod vendor
RUN debuild -us -uc

FROM debian:trixie
RUN --mount=type=bind,target=/tmp/build-executile,source=/tmp/build,from=build-executile \
    --mount=type=bind,target=/tmp/build-s3tftpd,source=/tmp/build,from=build-s3tftpd \
    apt-get update -qq && \
    apt-get install -y --no-install-recommends dumb-init /tmp/build-*/*.deb && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh

EXPOSE 69/udp
ENTRYPOINT ["dumb-init", "/docker-entrypoint.sh"]
