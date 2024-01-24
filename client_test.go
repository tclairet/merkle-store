package merklestore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"

	"github.com/tclairet/merklestore/merkletree"
)

type fakeFileHandler struct {
	saved string
}

func (f *fakeFileHandler) open(path string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader([]byte(path))), nil
}

func (f *fakeFileHandler) delete(path string) error {
	return nil
}

func (f *fakeFileHandler) save(name string, content []byte) error {
	f.saved = string(content)
	return nil
}

type fakeServer struct{}

func (f fakeServer) Upload(name string, file io.Reader) error {
	return nil
}

func TestUploader(t *testing.T) {
	fileHandler := &fakeFileHandler{}
	uploader := Uploader{
		server:      fakeServer{},
		builder:     merkletree.NewBuilder(),
		fileHandler: fileHandler,
	}
	if err := uploader.Upload([]string{"a", "b"}); err != nil {
		t.Fatal(err)
	}

	leaf1 := sha256.New()
	leaf1.Write([]byte("a"))
	h1 := leaf1.Sum(nil)

	leaf2 := sha256.New()
	leaf2.Write([]byte("b"))
	h2 := leaf2.Sum(nil)

	root := sha256.New()
	root.Write(h1)
	root.Write(h2)

	if got, want := fileHandler.saved, hex.EncodeToString(root.Sum(nil)); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
