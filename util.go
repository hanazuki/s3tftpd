package main

import (
	"errors"
	"net/url"
)

func normalizeKey(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] == '/' {
		s = s[1:]
	}
	if len(s) == 0 {
		return s
	}
	if s[len(s)-1] == '/' {
		s = s[0 : len(s)-1]
	}
	return s
}

func prefixKey(prefix, key string) string {
	if prefix == "" || prefix == "/" {
		return normalizeKey(key)
	}
	return normalizeKey(prefix) + "/" + normalizeKey(key)
}

func parseS3uri(uri url.URL) (bucket, prefix string, err error) {
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
