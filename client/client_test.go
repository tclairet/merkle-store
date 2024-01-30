package client

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"strconv"
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
	if rootFileName == name {
		return io.NopCloser(bytes.NewReader([]byte("{}"))), nil
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
	store   map[string][]byte
	tree    map[string]*merkletree.MerkleTree
	builder map[string]*merkletree.IndexedBuilder
}

func (f *fakeServer) Upload(root string, index int, total int, file io.Reader) error {
	b, _ := io.ReadAll(file)
	f.store[fmt.Sprintf("%s%d", root, index)] = b
	if _, exist := f.builder[root]; !exist {
		f.builder[root] = merkletree.NewIndexedBuilder(total)
	}
	hasher := sha256.New()
	hasher.Write(b)
	done, err := f.builder[root].AddHash(index, hasher.Sum(nil))
	if done {
		f.tree[root], _ = f.builder[root].Build()
	}
	return err
}

func (f *fakeServer) Request(root string, index int) (io.Reader, *merkletree.Proof, error) {
	stored := f.store[fmt.Sprintf("%s%d", root, index)]
	hasher := sha256.New()
	hasher.Write(stored)
	h1 := hasher.Sum(nil)
	hasher.Reset()
	hasher.Write([]byte(strconv.Itoa(index)))
	hasher.Write(h1)
	proof, err := f.tree[root].ProofFor(hasher.Sum(nil))
	if err != nil {
		return nil, nil, err
	}
	return bytes.NewReader(f.store[fmt.Sprintf("%s%d", root, index)]), proof, nil
}

func TestUploader(t *testing.T) {
	server := &fakeServer{
		store:   make(map[string][]byte),
		tree:    make(map[string]*merkletree.MerkleTree),
		builder: make(map[string]*merkletree.IndexedBuilder),
	}
	uploader := Uploader{
		server: server,
		fileHandler: &fakeFileHandler{
			saved: make(map[string][]byte),
		},
	}
	root, err := uploader.Upload([]string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}

	hasher := sha256.New()
	hasher.Write([]byte("a"))
	h1 := hasher.Sum(nil)
	hasher.Reset()
	hasher.Write([]byte("0"))
	hasher.Write(h1)
	h1 = hasher.Sum(nil)

	hasher.Reset()
	hasher.Write([]byte("b"))
	h2 := hasher.Sum(nil)
	hasher.Reset()
	hasher.Write([]byte("1"))
	hasher.Write(h2)
	h2 = hasher.Sum(nil)

	hasher.Reset()
	hasher.Write(h1)
	hasher.Write(h2)
	expectedRoot := hasher.Sum(nil)

	if got, want := root, hex.EncodeToString(expectedRoot); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	if err := uploader.Download(root, 0); err != nil {
		t.Error(err)
	}
	if err := uploader.Download(root, 1); err != nil {
		t.Error(err)
	}
}
