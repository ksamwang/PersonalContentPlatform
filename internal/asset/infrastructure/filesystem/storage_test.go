package filesystem

import (
	"bytes"
	"context"
	"io"
	"testing"
)

func TestPutOpenAndStat(t *testing.T) {
	s, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	body := []byte("asset payload")
	if err = s.Put(context.Background(), "original/ws/file", bytes.NewReader(body), int64(len(body)), "text/plain"); err != nil {
		t.Fatal(err)
	}
	info, err := s.Stat(context.Background(), "original/ws/file")
	if err != nil || info.Size != int64(len(body)) {
		t.Fatalf("stat: %#v %v", info, err)
	}
	reader, err := s.Open(context.Background(), "original/ws/file")
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	actual, _ := io.ReadAll(reader)
	if !bytes.Equal(actual, body) {
		t.Fatalf("got %q", actual)
	}
}
func TestRejectTraversal(t *testing.T) {
	s, _ := New(t.TempDir())
	if err := s.Put(context.Background(), "../escape", bytes.NewReader(nil), 0, ""); err == nil {
		t.Fatal("expected traversal rejection")
	}
}
