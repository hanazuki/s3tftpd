# s3tftpd

`s3tftpd` serves files on Amazon S3 via TFTP protocol, supporting both RRQ and WRQ.

## Usage

```
systemd-socket-activate -d -l 69 s3tftpd s3://your-bucket-name/prefix
```

Refer to `debian/s3tftpd.{socket,service}` for daemonization.
