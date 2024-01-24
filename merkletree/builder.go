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
	hash := h.Sum(nil)
	if len(hash) != 32 {
		return fmt.Errorf("wrong hash len")
	}
	builder.data = append(builder.data, hash)
	return nil
}

func (builder *Builder) Build() (*MerkleTree, error) {
	return FromHashes(builder.data)
}
