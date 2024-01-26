package client

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/tclairet/merklestore/merkletree"
)

type fakeFileHandler struct {
	saved map[string][]byte
}

func (f *fakeFileHandler) Open(name string) (io.ReadCloser, error) {
	if _, exist := f.saved[name]; exist {
		return io.NopCloser(bytes.NewReader(f.saved[name])), nil
	}
	return io.NopCloser(bytes.NewReader([]byte(name))), nil
}

func (f *fakeFileHandler) Delete(path string) error {
	return nil
}

func (f *fakeFileHandler) Save(name string, content io.Reader) error {
	b, _ := io.ReadAll(content)
	f.saved[name] = b
	return nil
}

type fakeServer struct {
	tree  *merkletree.MerkleTree
	store map[string][]byte
}

func (f fakeServer) Upload(root string, index int, total int, file io.Reader) error {
	b, _ := io.ReadAll(file)
	f.store[fmt.Sprintf("%s%d", root, index)] = b
	return nil
}

func (f fakeServer) Request(root string, index int) (io.Reader, *merkletree.Proof, error) {
	hasher := sha256.New()
	_, _ = io.Copy(hasher, bytes.NewReader(f.store[fmt.Sprintf("%s%d", root, index)]))
	proof, err := f.tree.ProofFor(hasher.Sum(nil))
	if err != nil {
		return nil, nil, err
	}
	return bytes.NewReader(f.store[fmt.Sprintf("%s%d", root, index)]), proof, nil
}

func TestUploader(t *testing.T) {
	server := &fakeServer{
		store: make(map[string][]byte),
	}
	uploader := Uploader{
		server:  server,
		builder: merkletree.NewBuilder(),
		fileHandler: &fakeFileHandler{
			saved: make(map[string][]byte),
		},
	}
	root, err := uploader.Upload([]string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	server.tree, _ = uploader.builder.Build()

	leaf1 := sha256.New()
	leaf1.Write([]byte("a"))
	h1 := leaf1.Sum(nil)

	leaf2 := sha256.New()
	leaf2.Write([]byte("b"))
	h2 := leaf2.Sum(nil)

	expectedRoot := sha256.New()
	expectedRoot.Write(h1)
	expectedRoot.Write(h2)

	if got, want := root, expectedRoot.Sum(nil); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if err := uploader.Download(root, 0); err != nil {
		t.Error(err)
	}
	if err := uploader.Download(root, 1); err != nil {
		t.Error(err)
	}
}
