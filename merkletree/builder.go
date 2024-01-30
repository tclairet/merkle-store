package merkletree

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"strconv"
)

type Builder struct {
	data [][]byte
}

func NewBuilder() *Builder {
	return &Builder{
		data: [][]byte{},
	}
}

func (builder *Builder) Add(input io.Reader) error {
	h := sha256.New()
	if _, err := io.Copy(h, input); err != nil {
		return err
	}
	builder.data = append(builder.data, h.Sum(nil))
	return nil
}

func (builder *Builder) Build() (*MerkleTree, error) {
	return FromHashes(builder.data, sha256.New)
}

type IndexedBuilder struct {
	data    [][]byte
	count   int
	newHash func() hash.Hash
}

func NewIndexedBuilder(size int) *IndexedBuilder {
	return &IndexedBuilder{
		data:    make([][]byte, size),
		newHash: sha256.New,
	}
}

func (builder *IndexedBuilder) Add(index int, input io.Reader) (bool, error) {
	hasher := builder.newHash()
	if _, err := io.Copy(hasher, input); err != nil {
		return false, err
	}
	if len(builder.data[index]) != 0 {
		return false, fmt.Errorf("already got hash for this index")
	}
	return builder.AddHash(index, hasher.Sum(nil))
}

func (builder *IndexedBuilder) AddHash(index int, h []byte) (bool, error) {
	if len(builder.data[index]) != 0 {
		return false, fmt.Errorf("already got hash for this index")
	}
	hasher := builder.newHash()
	hasher.Write([]byte(strconv.Itoa(index)))
	hasher.Write(h)

	builder.data[index] = hasher.Sum(nil)
	builder.count++

	return builder.count == len(builder.data), nil
}

func (builder *IndexedBuilder) Build() (*MerkleTree, error) {
	return FromHashes(builder.data, builder.newHash)
}
