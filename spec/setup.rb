#!/usr/bin/env ruby
require 'aws-sdk-s3'

options = {region: ENV.fetch('AWS_REGION', ENV['AWS_DEFAULT_REGION'])}
if endpoint_url = ENV['S3TFTPD_TEST_ENDPOINT_URL']
  options[:endpoint] = endpoint_url
  options[:force_path_style] = true
end

s3 = Aws::S3::Resource.new(**options)
bucket_name = ENV.fetch('S3TFTPD_TEST_BUCKET_NAME')
bucket = s3.bucket(bucket_name)

bucket.create unless bucket.exists?

bucket.policy.delete

bucket.object('test1').put(body: "test object 1\n")
bucket.object('prefix/test2').put(body: "test object 2\n")

bucket.policy.put({
  policy: JSON.dump(
    {
      Version: "2012-10-17",
      Statement: [
        {
          Effect: "Deny",
          Principal: "*",
          Action: %w[s3:PutObject s3:PutObjectAcl],
          NotResource: "arn:aws:s3:::#{bucket_name}/writable/*"
        }
      ]
    }
  )
})

bucket.lifecycle.put({
  lifecycle_configuration: {
    rules: [
      {
        expiration: {
          days: 1,
        },
        id: "DeleteTempFiles",
        prefix: "writable/",
        status: "Enabled",
        abort_incomplete_multipart_upload: {
          days_after_initiation: 1,
        },
      },
    ],
  },
})
