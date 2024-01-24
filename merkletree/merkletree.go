package merkletree

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
)

type MerkleTree struct {
	nodes  map[string]Node
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

func FromHashes(hashes [][]byte) (*MerkleTree, error) {
	tree := &MerkleTree{
		newHash: sha256.New,
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

func (tree *MerkleTree) build(hashes [][]byte) error {
	if len(hashes) == 0 {
		return fmt.Errorf("invalid inputs")
	}

	tree.height = 1
	tree.nodes = make(map[string]Node)

	for _, hash := range hashes {
		tree.nodes[hex.EncodeToString(hash)] = Node{
			hash: hash,
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
		h := tree.newHash()
		h.Write(hashes[i*2])
		h.Write(hashes[i*2+1])
		hash := h.Sum(nil)
		tree.nodes[hex.EncodeToString(hash)] = Node{
			hash:       hash,
			leftChild:  hashes[i*2],
			rightChild: hashes[i*2+1],
		}
		newNodes = append(newNodes, hash)
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
	hash       []byte
	leftChild  []byte
	rightChild []byte
}

func isOdd(number int) bool {
	return number%2 == 1
}
