package merkletree

import (
	"crypto/sha256"
	"fmt"
	"io"
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
	return FromHashes(builder.data)
}

type IndexedBuilder struct {
	data  [][]byte
	count int
}

func NewIndexedBuilder(size int) *IndexedBuilder {
	return &IndexedBuilder{
		data: make([][]byte, size),
	}
}

func (builder *IndexedBuilder) Add(index int, input io.Reader) (bool, error) {
	h := sha256.New()
	if _, err := io.Copy(h, input); err != nil {
		return false, err
	}
	if len(builder.data[index]) != 0 {
		return false, fmt.Errorf("already got hash for this index")
	}
	builder.data[index] = h.Sum(nil)
	builder.count++
	return builder.count == len(builder.data), nil
}

func (builder *IndexedBuilder) AddHash(index int, h []byte) (bool, error) {
	if len(builder.data[index]) != 0 {
		return false, fmt.Errorf("already got hash for this index")
	}
	builder.data[index] = h
	builder.count++
	return builder.count == len(builder.data), nil
}

func (builder *IndexedBuilder) Build() (*MerkleTree, error) {
	return FromHashes(builder.data)
}
