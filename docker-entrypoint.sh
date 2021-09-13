#!/bin/bash
set -eu

: ${S3TFTPD_LISTEN_PORT:=69}

if [[ -d /docker-entrypoint.d ]]; then
    run-parts --exit-on-error /docker-entrypoint.d
fi

if [[ ${1-} == /* ]]; then
    exec "$@"
fi

if [[ ${1-} == s3tftpd ]]; then
    shift
fi

exec /usr/bin/inet-socket-listen --udp --name=tftp --numeric-host ::0 "${S3TFTPD_LISTEN_PORT}" s3tftpd --single-port "$@"
