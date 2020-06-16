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

args=(-d -l "${S3TFTPD_LISTEN_PORT}")
while IFS= read -r -d $'\0' env; do
    env="$(cut -d= -f1 <<<"$env")"
    if [[ "$env" == AWS_* ]]; then
        args+=(-E "$env")
    fi
done < <(printenv -0)

exec /usr/bin/systemd-socket-activate "${args[@]}" s3tftpd --single-port "$@"
