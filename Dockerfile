# syntax = docker/dockerfile:experimental

FROM debian:buster as builder
RUN apt-get update -qq && apt-get install -y --no-install-recommends devscripts equivs git

WORKDIR /tmp/build/src
COPY debian/control debian/
RUN yes | mk-build-deps -i
COPY . .
RUN --mount=type=cache,target=/root/go/pkg/mod go mod vendor
RUN debuild -us -uc

FROM debian:buster
RUN --mount=type=bind,target=/tmp/build,source=/tmp/build,from=builder \
    apt-get update -qq && \
    apt-get install -y --no-install-recommends systemd /tmp/build/*.deb && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh

EXPOSE 69/udp
ENTRYPOINT ["dumb-init", "/docker-entrypoint.sh"]
