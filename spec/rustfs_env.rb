#!/usr/bin/env ruby
# Setup environment variable for RustFS integration test.
# Usage: ./spec/rustfs_env.rb bundle exec rake spec

require 'json'

command = ARGV
abort 'No command specified' if command.empty?
command[0] = command[0].yield_self { [it, it] }

compose_file = File.join(__dir__, 'docker-compose.yml')

port = IO.popen(['docker', 'compose', '-f', compose_file, 'port', 'rustfs', '9000'], &:read).chomp.split(':').last
abort 'Failed to discover RustFS port.' if !port || port.empty?

config = JSON.parse(IO.popen(['docker', 'compose', '-f', compose_file, 'config', '--format', 'json'], &:read))
compose_env = config.dig('services', 'rustfs', 'environment')

env = {
  "AWS_ACCESS_KEY_ID" => compose_env.fetch('RUSTFS_ROOT_USER'),
  "AWS_SECRET_ACCESS_KEY" => compose_env.fetch('RUSTFS_ROOT_PASSWORD'),
  "AWS_DEFAULT_REGION" => "us-east-1",
  "S3TFTPD_TEST_ENDPOINT_URL" => "http://localhost:#{port}",
  "S3TFTPD_TEST_BUCKET_NAME" => "s3tftpd-test",
}

exec(env, *command)
