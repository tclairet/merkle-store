package merkletree

import (
	"bytes"
	"fmt"
	"hash"
)

type Proof struct {
	newHash func() hash.Hash
	hashes  [][]byte
}

func NewProof(newHash func() hash.Hash, hashes [][]byte) *Proof {
	return &Proof{
		hashes:  hashes,
		newHash: newHash,
	}
}

func (proof Proof) Verify(leaf []byte, root []byte) error {
	if !bytes.Equal(leaf, proof.hashes[0]) {
		return fmt.Errorf("invalid start leaf")
	}
	if len(proof.hashes) == 1 {
		return nil
	}

	hash := leaf
	for i := 1; i < len(proof.hashes); i = i + 2 {
		if err := proof.validate(hash, proof.hashes[i], proof.hashes[i+1]); err != nil {
			return err
		}
		hash = proof.hashes[i+1]
	}
	if !bytes.Equal(hash, root) {
		return fmt.Errorf("root mismatch, got %x want %x", hash, root)
	}
	return nil
}

func (proof Proof) validate(left, right, expected []byte) error {
	hash1 := sum(proof.newHash, left, right)
	hash2 := sum(proof.newHash, right, left)
	if !bytes.Equal(hash1, expected) && !bytes.Equal(hash2, expected) {
		return fmt.Errorf("cannot verify, calculated %x and %x but expected %x", hash1, hash2, expected)
	}
	return nil
}
