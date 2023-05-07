package main

import "testing"

func TestPrefixKey(t *testing.T) {
	if actual, expected := prefixKey("a", "b"), "a/b"; actual != expected {
		t.Errorf("expected %v, actual %v", expected, actual)
	}

	if actual, expected := prefixKey("a/", "b"), "a/b"; actual != expected {
		t.Errorf("expected %v, actual %v", expected, actual)
	}

	if actual, expected := prefixKey("a/", "/b"), "a/b"; actual != expected {
		t.Errorf("expected %v, actual %v", expected, actual)
	}

	if actual, expected := prefixKey("/a", "b"), "a/b"; actual != expected {
		t.Errorf("expected %v, actual %v", expected, actual)
	}

	if actual, expected := prefixKey("/a/", "b"), "a/b"; actual != expected {
		t.Errorf("expected %v, actual %v", expected, actual)
	}

	if actual, expected := prefixKey("/a/", "/b"), "a/b"; actual != expected {
		t.Errorf("expected %v, actual %v", expected, actual)
	}

	if actual, expected := prefixKey("", "/b"), "b"; actual != expected {
		t.Errorf("expected %v, actual %v", expected, actual)
	}

	if actual, expected := prefixKey("/", "/b"), "b"; actual != expected {
		t.Errorf("expected %v, actual %v", expected, actual)
	}
}
