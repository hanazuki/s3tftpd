# Hacking on s3tftpd

## Prerequisites

- Go toolchain (see `go.mod` for minimum version)
- Ruby with Bundler (`bundle install` to install test dependencies)
- asciidoctor (for building man pages via `make man`)
- tftp-hpa (for integration tests)
- Docker and Docker Compose v2 (for integration tests against RustFS)

## Running integration tests against RustFS

```shell
# Start RustFS:
docker compose -f spec/docker-compose.yml up -d --wait

# Seed the test bucket:
./spec/rustfs_env.rb bundle exec ruby spec/setup.rb

# Run the tests:
./spec/rustfs_env.rb bundle exec rspec

# When done, stop RustFS:
docker compose -f spec/docker-compose.yml down
```

## Running integration tests against real AWS S3

Configure AWS credentials via the standard AWS SDK mechanisms (environment variables, `~/.aws/credentials`, etc.), then:

```sh
export AWS_REGION=us-east-2
export S3TFTPD_TEST_BUCKET_NAME=your-bucket-name

# Seed the bucket:
bundle exec ruby spec/setup.rb

# Run the tests:
bundle exec rspec
```
