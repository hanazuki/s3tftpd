package main

import (
	"errors"
	"net/url"
)

type S3Uri struct {
	Bucket, Prefix string
}

func (s *S3Uri) UnmarshalText(text []byte) error {
	uri, err := url.Parse(string(text))
	if err != nil {
		return err
	}

	if uri.Scheme != "s3" {
		return errors.New("S3URI must have 's3' scheme")
	}
	if uri.Host == "" {
		return errors.New("S3URI must contain bucket name")
	}

	s.Bucket = uri.Host
	s.Prefix = normalizeKey(uri.Path)
	if s.Prefix != "" {
		s.Prefix += "/"
	}

	return nil
}

func (s *S3Uri) GetKey(suffix string) string {
	return s.Prefix + normalizeKey(suffix)
}

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
