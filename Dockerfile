# syntax = docker/dockerfile:experimental

FROM debian:buster as builder
# https://salsa.debian.org/go-team/packages/dh-golang/-/commit/61b0829ad608be1aa23630e9d8f9d76ded3eca65
RUN echo "deb http://deb.debian.org/debian buster-backports main" > /etc/apt/sources.list.d/backports.list
RUN (echo "Package: dh-*"; echo "Pin: release a=buster-backports"; echo "Pin-Priority: 500") > /etc/apt/preferences.d/99debhelper
RUN apt-get update -qq && apt-get install -y --no-install-recommends devscripts equivs git

WORKDIR /tmp/build/src
COPY debian/control debian/
RUN yes | mk-build-deps -i
COPY . .
RUN --mount=type=cache,target=/root/go/pkg/mod go mod vendor
RUN debuild -us -uc

FROM debian:buster
ARG SOURCE_COMMIT

RUN --mount=type=bind,target=/tmp/build,source=/tmp/build,from=builder \
    apt-get update -qq && \
    apt-get install -y --no-install-recommends dumb-init systemd /tmp/build/*.deb && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh

EXPOSE 69/udp
ENTRYPOINT ["dumb-init", "/docker-entrypoint.sh"]

RUN echo "$SOURCE_COMMIT" > /REVISION
