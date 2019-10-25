export GO111MODULE=on

all:
	go build

install:
	install -D -t ${DESTDIR}/usr/sbin s3tftpd

clean:
	rm -f s3tftpd

.PHONY: all install clean
