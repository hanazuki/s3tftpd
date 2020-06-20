export GO111MODULE=on

GOCMD = go
RM = rm
INSTALL = install
DESTDIR =

all: build man

build:
	${GOCMD} build

man:
	asciidoctor -b manpage man/*.adoc

install:
	${INSTALL} -D -s -t ${DESTDIR}/usr/sbin s3tftpd
	${INSTALL} -D -t ${DESTDIR}/usr/share/man/man8 man/*.8

clean:
	${RM} -f s3tftpd
	${RM} -f man/*.8

.PHONY: all man install clean
