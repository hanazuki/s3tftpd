export GO111MODULE=on

GOCMD = go
RM = rm
INSTALL = install
DESTDIR =

all:
	${GOCMD} build
	asciidoctor -b manpage man/*.adoc

install:
	${INSTALL} -D -s -t ${DESTDIR}/usr/sbin s3tftpd
	${INSTALL} -D -t ${DESTDIR}/usr/share/man/man1 man/*.1

clean:
	${RM} -f s3tftpd
	${RM} -f man/*.1

.PHONY: all install clean
