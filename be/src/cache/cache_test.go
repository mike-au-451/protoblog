package cache

import "testing"

func TestPut(t *testing.T) {
	content1 := []byte("abcdefghijklmnopqrstuvwxyz")
	filename1 := "foo"
	content2 := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	filename2 := "bar"

	cc := New("/home/mike/Play/2023/06/20230625_01/be/testing", 0)
	if cc == nil {
		t.Fatalf("failed to create cache")
	}

	var err error

	err = cc.Put(filename1, content1)
	if err != nil {
		t.Fatalf("failed to Put 1")
	}
	err = cc.Put(filename2, content2)
	if err != nil {
		t.Fatalf("failed to Put 2")
	}
}
