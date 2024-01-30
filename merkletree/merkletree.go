package merkletree

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
)

type MerkleTree struct {
	nodes  map[string]*Node
	root   []byte
	height int

	newHash func() hash.Hash
}

func From(inputs [][]byte) (*MerkleTree, error) {
	tree := &MerkleTree{
		newHash: sha256.New,
	}
	if err := tree.from(inputs); err != nil {
		return nil, err
	}
	return tree, nil
}

func FromHashes(hashes [][]byte, newHash func() hash.Hash) (*MerkleTree, error) {
	tree := &MerkleTree{
		newHash: newHash,
	}
	if err := tree.build(hashes); err != nil {
		return nil, err
	}
	return tree, nil
}

func (tree *MerkleTree) Level(index int) ([][]byte, error) {
	if tree.height == 0 {
		return nil, fmt.Errorf("merkle tree not initialized")
	}

	if index >= tree.height {
		return nil, fmt.Errorf("cannot retrieve level '%d' height is '%d'", index, tree.height)
	}

	parents := [][]byte{tree.root}
	if index == 0 {
		return parents, nil
	}

	var nodes [][]byte
	for i := 0; i < index; i++ {
		nodes = [][]byte{}
		for _, child := range parents {
			nodes = append(nodes, tree.childOf(child)...)
		}
		parents = nodes
	}

	return nodes, nil
}

func (tree *MerkleTree) Height() int {
	return tree.height
}

func (tree *MerkleTree) Root() []byte {
	return tree.root
}

func (tree *MerkleTree) ProofFor(hash []byte) (*Proof, error) {
	node, exist := tree.nodes[hex.EncodeToString(hash)]
	if !exist {
		return nil, fmt.Errorf("%x not found in merkle tree", hash)
	}
	var hashes = [][]byte{node.hash}
	for !bytes.Equal(node.hash, tree.root) {
		node = tree.nodes[hex.EncodeToString(node.parent)]
		otherLeaf := node.rightChild
		if bytes.Equal(hashes[len(hashes)-1], otherLeaf) {
			otherLeaf = node.leftChild
		}
		hashes = append(hashes, otherLeaf, node.hash)
	}
	return NewProof(tree.newHash, hashes), nil
}

func (tree *MerkleTree) build(hashes [][]byte) error {
	if len(hashes) == 0 {
		return fmt.Errorf("invalid inputs")
	}

	tree.height = 1
	tree.nodes = make(map[string]*Node)

	for _, h := range hashes {
		tree.nodes[hex.EncodeToString(h)] = &Node{
			hash: h,
		}
	}

	for len(hashes) != 1 {
		hashes = tree.buildBranch(hashes)
	}
	tree.root = hashes[0]

	return nil
}

func (tree *MerkleTree) from(inputs [][]byte) error {
	var hashes [][]byte
	for _, data := range inputs {
		h := tree.newHash()
		h.Write(data)
		hashes = append(hashes, h.Sum(nil))
	}
	return tree.build(hashes)
}

func (tree *MerkleTree) buildBranch(hashes [][]byte) [][]byte {
	var newNodes [][]byte

	for i := 0; i < (len(hashes) / 2); i++ {
		h := sum(tree.newHash, hashes[i*2], hashes[i*2+1])
		tree.nodes[hex.EncodeToString(h)] = &Node{
			hash:       h,
			leftChild:  hashes[i*2],
			rightChild: hashes[i*2+1],
		}
		tree.nodes[hex.EncodeToString(hashes[i*2])].parent = h
		tree.nodes[hex.EncodeToString(hashes[i*2+1])].parent = h
		newNodes = append(newNodes, h)
	}

	if isOdd(len(hashes)) {
		newNodes = append(newNodes, hashes[len(hashes)-1])
	}

	tree.height++
	return newNodes
}

func (tree *MerkleTree) childOf(hash []byte) [][]byte {
	node := tree.nodes[hex.EncodeToString(hash)]
	if len(node.leftChild) == 0 && len(node.rightChild) == 0 {
		return [][]byte{hash}
	}
	return [][]byte{node.leftChild, node.rightChild}
}

type Node struct {
	parent     []byte
	hash       []byte
	leftChild  []byte
	rightChild []byte
}

func isOdd(number int) bool {
	return number%2 == 1
}

func sum(newHash func() hash.Hash, left, right []byte) []byte {
	h := newHash()
	h.Write(left)
	h.Write(right)
	return h.Sum(nil)
}
