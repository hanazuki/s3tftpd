FROM golang:1.14 as builder

WORKDIR /build
COPY . .

RUN go build

FROM debian:buster
RUN apt-get update -qq && apt-get install -y --no-install-recommends ca-certificates systemd && apt-get clean && rm -rf /var/lib/apt/lists/*
COPY --from=builder /build/s3tftpd /usr/local/bin
COPY debian/copyright /
COPY docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh

EXPOSE 69/udp
ENTRYPOINT ["/docker-entrypoint.sh"]
