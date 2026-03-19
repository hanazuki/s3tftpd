package main

import "testing"

func TestS3Uri(t *testing.T) {
	cases := []struct {
		input    string
		suffix   string
		expected string
	}{
		{"s3://bucket/a", "b", "a/b"},
		{"s3://bucket/a", "/b", "a/b"},
		{"s3://bucket/a/", "b", "a/b"},
		{"s3://bucket/a/", "/b", "a/b"},
		{"s3://bucket", "b", "b"},
		{"s3://bucket/", "b", "b"},
	}

	for _, c := range cases {
		var s S3Uri
		if err := s.UnmarshalText([]byte(c.input)); err != nil {
			t.Errorf("UnmarshalText(%q) error: %v", c.input, err)
			continue
		}
		if actual := s.GetKey(c.suffix); actual != c.expected {
			t.Errorf("GetKey(%q) on %q: expected %q, got %q", c.suffix, c.input, c.expected, actual)
		}
	}
}

func TestS3UriErrors(t *testing.T) {
	cases := []string{
		"http://bucket/a",
		"s3://",
	}

	for _, input := range cases {
		var s S3Uri
		if err := s.UnmarshalText([]byte(input)); err == nil {
			t.Errorf("UnmarshalText(%q) expected error, got nil", input)
		}
	}
}
