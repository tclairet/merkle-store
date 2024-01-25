package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/tclairet/merklestore/files"
)

type store interface {
	save(root string, hash []byte, index, total int) error
	get(root string, index int) ([]byte, error)
	read() map[string][][]byte
}

type memStore struct {
	Hashes map[string][][]byte `json:"names,omitempty"`

	mu sync.Mutex
}

func (mem *memStore) save(root string, hash []byte, index, total int) error {
	mem.mu.Lock()
	defer mem.mu.Unlock()
	if len(mem.Hashes[root]) == 0 {
		mem.Hashes[root] = make([][]byte, total)
	}
	mem.Hashes[root][index] = hash
	return nil
}

func (mem *memStore) get(root string, index int) ([]byte, error) {
	mem.mu.Lock()
	defer mem.mu.Unlock()
	if len(mem.Hashes[root]) < index {
		return nil, fmt.Errorf("invalid index")
	}
	return mem.Hashes[root][index], nil
}

func (mem *memStore) read() map[string][][]byte {
	return mem.Hashes
}

type JsonStore struct {
	*memStore
	files files.Handler
}

func NewJsonStore(files files.Handler) (*JsonStore, error) {
	store := memStore{
		Hashes: make(map[string][][]byte),
	}
	reader, err := files.Open("backup.json")
	if err == nil { // backup exist
		if err := json.NewDecoder(reader).Decode(&store); err != nil {
			return nil, err
		}
	}
	return &JsonStore{
		memStore: &store,
		files:    files,
	}, nil
}

func (store *JsonStore) save(root string, hash []byte, index, total int) error {
	if err := store.memStore.save(root, hash, index, total); err != nil {
		return err
	}
	b, err := json.Marshal(store.Hashes)
	if err != nil {
		return err
	}
	return store.files.Save("backup.json", bytes.NewBuffer(b))
}
