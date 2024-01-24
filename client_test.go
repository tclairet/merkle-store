package merklestore

import (
	"bytes"
	"crypto/sha256"
	"io"
	"reflect"
	"testing"

	"github.com/tclairet/merklestore/merkletree"
)

type fakeFileHandler struct {
	saved map[string][]byte
}

func (f *fakeFileHandler) open(name string) (io.ReadCloser, error) {
	if _, exist := f.saved[name]; exist {
		return io.NopCloser(bytes.NewReader(f.saved[name])), nil
	}
	return io.NopCloser(bytes.NewReader([]byte(name))), nil
}

func (f *fakeFileHandler) delete(path string) error {
	return nil
}

func (f *fakeFileHandler) save(name string, content []byte) error {
	f.saved[name] = content
	return nil
}

type fakeServer struct {
	tree  *merkletree.MerkleTree
	store map[string][]byte
}

func (f fakeServer) Upload(name string, file io.Reader) error {
	b, _ := io.ReadAll(file)
	f.store[name] = b
	return nil
}

func (f fakeServer) Request(name string) (io.Reader, *merkletree.Proof, error) {
	hasher := sha256.New()
	_, _ = io.Copy(hasher, bytes.NewReader(f.store[name]))
	proof, err := f.tree.ProofFor(hasher.Sum(nil))
	if err != nil {
		return nil, nil, err
	}
	return bytes.NewReader(f.store[name]), proof, nil
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
	if err := uploader.Upload([]string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	server.tree, _ = uploader.builder.Build()

	leaf1 := sha256.New()
	leaf1.Write([]byte("a"))
	h1 := leaf1.Sum(nil)

	leaf2 := sha256.New()
	leaf2.Write([]byte("b"))
	h2 := leaf2.Sum(nil)

	root := sha256.New()
	root.Write(h1)
	root.Write(h2)

	gotRoot, _ := uploader.getRoot()
	if got, want := gotRoot, root.Sum(nil); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if err := uploader.Download("a"); err != nil {
		t.Error(err)
	}
}
