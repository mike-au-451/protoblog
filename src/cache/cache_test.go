package cache

import (
	"bytes"
	"testing"
)

func TestGet(t *testing.T) {
	type testpair struct {
		key string
		expect []byte
		ok bool
	}

	tests := []testpair {
		{ "test_aa", []byte("aa"), true },
		{ "test_bb", []byte("bb"), true },
		{ "test_cc", []byte("cc"), true },
		{ "test_dd", []byte(""), false },
	}

	c := New(10)

	for _, test := range tests {
		got, ok := c.Get(test.key)
		if test.ok != ok {
			t.Fatalf("%s: unexpected ok, expected %v, got %v", test.key, test.ok, ok)
		}
		if !test.ok {
			continue
		}
		if bytes.Compare(test.expect, got) != 0 {
			t.Fatalf("%s: expected %s, got %s", test.key, test.expect, got)
		}
	}
}