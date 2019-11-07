# s3tftpd

`s3tftpd` serves files on Amazon S3 via TFTP protocol, supporting both RRQ and WRQ.

## Usage

```
systemd-socket-activate -d -l 69 s3tftpd s3://your-bucket-name/prefix/
```

Refer to `s3tftpd --help` for command line options and `debian/s3tftpd.{socket,service}` for daemonization.

`s3tftpd` retrieves AWS credentials from the `AWS_*` environment variables, shared profile file or EC2/ECS role.
Because of the nature of TFTP `s3tftpd` has no mechanisms of client authentication. Access controls on the objects should be enforced using IAM Policies and S3 Bucket Policies.


## Docker container

Prebuilt container images are available at [Docker Hub](https://hub.docker.com/r/hanazuki/s3tftpd). Versioned tags are available and `latest` is an alias for the latest tagged release.
