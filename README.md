# s3tftpd

`s3tftpd` serves files on Amazon S3 via TFTP protocol, supporting both RRQ and WRQ.

## Usage

`s3tftpd` expects to receive a file descriptor for a UDP socket from the system manager (e.g. systemd). If you are not using a compatible system manager, you can use [`systemd-socket-activate(1)`](https://www.freedesktop.org/software/systemd/man/systemd-socket-activate.html) to pass an FD to `s3tftpd`.

```
systemd-socket-activate -d -l 69 s3tftpd s3://your-bucket-name/prefix/
```

Refer to `s3tftpd --help` for command line options and `debian/s3tftpd.{socket,service}` for daemonization.

`s3tftpd` retrieves AWS credentials from the `AWS_*` environment variables, shared profile file or EC2/ECS role.
Because of the nature of TFTP `s3tftpd` has no mechanisms of client authentication. Access controls on the objects should be enforced using IAM Policies and S3 Bucket Policies.


## Docker container

Prebuilt container images are available at [Docker Hub](https://hub.docker.com/r/hanazuki/s3tftpd). Available tags are `latest` (the latest release), `testing` (master branch), and each versioned release.
