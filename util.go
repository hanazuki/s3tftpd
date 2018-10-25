package main

import (
	"errors"
	"net/url"
)

func normalizeKey(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] != '/' {
		s = "/" + s
	}
	if s[len(s)-1] == '/' {
		s = s[0 : len(s)-1]
	}
	return s
}

func prefixKey(prefix, key string) string {
	return normalizeKey(prefix) + normalizeKey(key)
}

func parseS3uri(rawuri string) (bucket, prefix string, err error) {
	uri, err := url.Parse(rawuri)
	if err != nil {
		return
	}

	if uri.Scheme != "s3" {
		err = errors.New("S3URI must have 's3' scheme")
		return
	}
	if uri.Host == "" {
		err = errors.New("S3URI must contain bucket name")
		return
	}

	bucket = uri.Host
	prefix = uri.Path

	return
}
