#!/bin/bash
set -eu

: ${S3TFTPD_LISTEN_PORT:=69}

if [[ -d /docker-entrypoint.d ]]; then
    run-parts --exit-on-error /docker-entrypoint.d
fi

if [[ ${1-} == /* ]]; then
    exec "$@"
fi

exec /usr/bin/systemd-socket-activate -d -l "${S3TFTPD_LISTEN_PORT}" s3tftpd "$@"
