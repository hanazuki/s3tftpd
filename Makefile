export GO111MODULE=on

GOCMD = go
RM = rm
INSTALL = install
DESTDIR =

all:
	${GOCMD} build

install:
	${INSTALL} -D -t ${DESTDIR}/usr/sbin s3tftpd

clean:
	${RM} -f s3tftpd

.PHONY: all install clean
