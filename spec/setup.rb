#!/usr/bin/env ruby
require 'aws-sdk-s3'

s3 = Aws::S3::Resource.new
bucket_name = ENV.fetch('TEST_BUCKET_NAME')
bucket = s3.bucket(bucket_name)

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
          Action: "s3:PutObject*",
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
