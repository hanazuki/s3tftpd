FROM golang:1.11 as builder

WORKDIR /go/src/github.com/hanazuki/s3tftpd
COPY . .

ENV GO111MODULE=on
RUN go build

FROM debian:stretch
RUN apt-get update -qq && apt-get install -y --no-install-recommends systemd && apt-get clean && rm -rf /var/lib/apt/lists/*
COPY --from=builder /go/src/github.com/hanazuki/s3tftpd/s3tftpd /usr/local/bin
COPY docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh

EXPOSE 69/udp
ENTRYPOINT ["/docker-entrypoint.sh"]
