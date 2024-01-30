package merkletree

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// fakeHash only append provided bytes without hashing anything
type fakeHash struct {
	bytes []byte
}

func (f *fakeHash) Write(p []byte) (n int, err error) {
	f.bytes = append(f.bytes, p...)
	return len(p), nil
}

func (f *fakeHash) Sum(b []byte) []byte {
	f.bytes = append(f.bytes, b...)
	return f.bytes
}

func (f *fakeHash) Reset() {
	f.bytes = []byte{}
}

func (f *fakeHash) Size() int {
	return 0
}

func (f *fakeHash) BlockSize() int {
	return 0
}

func newFakeHash() hash.Hash {
	return &fakeHash{bytes: make([]byte, 0)}
}

func TestMerkleTree(t *testing.T) {
	t.Run("from", func(t *testing.T) {
		tree := MerkleTree{newHash: newFakeHash}

		cases := []struct {
			inputs         []string
			expectedRoot   string
			expectedHeight int
		}{
			{inputs: []string{"a"}, expectedRoot: "a", expectedHeight: 1},
			{inputs: []string{"a", "b"}, expectedRoot: "ab", expectedHeight: 2},
			{inputs: []string{"a", "b", "c"}, expectedRoot: "abc", expectedHeight: 3},
			{inputs: []string{"a", "b", "c", "d"}, expectedRoot: "abcd", expectedHeight: 3},
			{inputs: []string{"a", "b", "c", "d", "e"}, expectedRoot: "abcde", expectedHeight: 4},
			{inputs: []string{"a", "b", "c", "d", "e", "f"}, expectedRoot: "abcdef", expectedHeight: 4},
			{inputs: []string{"a", "b", "c", "d", "e", "f", "g", "h"}, expectedRoot: "abcdefgh", expectedHeight: 4},
			{inputs: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}, expectedRoot: "abcdefghi", expectedHeight: 5},
		}

		for _, c := range cases {
			t.Run(strings.Join(c.inputs, ""), func(t *testing.T) {
				tree.from(stringsToBytes(c.inputs))

				if got, want := string(tree.root), c.expectedRoot; got != want {
					t.Fatalf("got %v, want %v", got, want)
				}
				if got, want := tree.Height(), c.expectedHeight; got != want {
					t.Fatalf("got %v, want %v", got, want)
				}
			})
		}
	})

	t.Run("level", func(t *testing.T) {
		tree := MerkleTree{newHash: newFakeHash}

		cases := []struct {
			inputs   []string
			level    int
			expected []string
		}{
			{[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}, 0, []string{"abcdefghi"}},
			{[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}, 1, []string{"abcdefgh", "i"}},
			{[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}, 2, []string{"abcd", "efgh", "i"}},
			{[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}, 3, []string{"ab", "cd", "ef", "gh", "i"}},
			{[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}, 4, []string{"a", "b", "c", "d", "e", "f", "g", "h", "i"}},
		}

		for _, c := range cases {
			t.Run(fmt.Sprintf("%d", c.level), func(t *testing.T) {
				tree.from(stringsToBytes(c.inputs))

				leaves, err := tree.Level(c.level)
				if err != nil {
					t.Fatal(err)
				}

				if got, want := leaves, stringsToBytes(c.expected); !reflect.DeepEqual(got, want) {
					t.Fatalf("got %v, want %v", got, want)
				}
			})
		}
	})

	t.Run("invalid level", func(t *testing.T) {
		tree, _ := From(stringsToBytes([]string{"a", "b"}))
		cases := []struct {
			tree     *MerkleTree
			level    int
			expected string
		}{
			{&MerkleTree{}, 0, "merkle tree not initialized"},
			{tree, 2, "cannot retrieve level"},
		}

		for _, c := range cases {
			_, err := c.tree.Level(c.level)
			if got, want := err.Error(), c.expected; !strings.Contains(got, want) {
				t.Fatalf("'%v', does not contains %v", got, want)
			}
		}
	})

	t.Run("Proof", func(t *testing.T) {

		cases := []struct {
			inputs        []string
			index         int
			expectedProof []string
		}{
			{inputs: []string{"a"}, index: 0, expectedProof: []string{"0a"}},
			{inputs: []string{"a", "b"}, index: 0, expectedProof: []string{"0a", "1b", "0a1b"}},
			{inputs: []string{"a", "b", "c"}, index: 0, expectedProof: []string{"0a", "1b", "0a1b", "2c", "0a1b2c"}},
			{inputs: []string{"a", "b", "c"}, index: 2, expectedProof: []string{"2c", "0a1b", "0a1b2c"}},
			{inputs: []string{"a", "b", "c", "d"}, index: 0, expectedProof: []string{"0a", "1b", "0a1b", "2c3d", "0a1b2c3d"}},
			{inputs: []string{"a", "b", "c", "d"}, index: 1, expectedProof: []string{"1b", "0a", "0a1b", "2c3d", "0a1b2c3d"}},
			{inputs: []string{"a", "b", "c", "d"}, index: 2, expectedProof: []string{"2c", "3d", "2c3d", "0a1b", "0a1b2c3d"}},
			{inputs: []string{"a", "b", "c", "d"}, index: 3, expectedProof: []string{"3d", "2c", "2c3d", "0a1b", "0a1b2c3d"}},
			{inputs: []string{"a", "b", "c", "d", "e"}, index: 0, expectedProof: []string{"0a", "1b", "0a1b", "2c3d", "0a1b2c3d", "4e", "0a1b2c3d4e"}},
			{inputs: []string{"a", "b", "c", "d", "e"}, index: 4, expectedProof: []string{"4e", "0a1b2c3d", "0a1b2c3d4e"}},
			{inputs: []string{"a", "b", "c", "d", "e", "f"}, index: 0, expectedProof: []string{"0a", "1b", "0a1b", "2c3d", "0a1b2c3d", "4e5f", "0a1b2c3d4e5f"}},
			{inputs: []string{"a", "b", "c", "d", "e", "f", "g"}, index: 0, expectedProof: []string{"0a", "1b", "0a1b", "2c3d", "0a1b2c3d", "4e5f6g", "0a1b2c3d4e5f6g"}},
			{inputs: []string{"a", "b", "c", "d", "e", "f", "g", "h"}, index: 0, expectedProof: []string{"0a", "1b", "0a1b", "2c3d", "0a1b2c3d", "4e5f6g7h", "0a1b2c3d4e5f6g7h"}},
			{inputs: []string{"a", "b", "c", "d", "e", "f", "g", "h"}, index: 3, expectedProof: []string{"3d", "2c", "2c3d", "0a1b", "0a1b2c3d", "4e5f6g7h", "0a1b2c3d4e5f6g7h"}},
			{inputs: []string{"a", "b", "c", "d", "e", "f", "g", "h"}, index: 7, expectedProof: []string{"7h", "6g", "6g7h", "4e5f", "4e5f6g7h", "0a1b2c3d", "0a1b2c3d4e5f6g7h"}},
		}

		for _, c := range cases {
			t.Run(fmt.Sprintf("%s for %s", strings.Join(c.inputs, ""), c.inputs[c.index]), func(t *testing.T) {
				builder := NewIndexedBuilder(len(c.inputs))
				builder.newHash = newFakeHash
				for i, input := range c.inputs {
					builder.AddHash(i, []byte(input))
				}
				tree, err := builder.Build()
				if err != nil {
					t.Fatal(err)
				}

				indexedHash := []byte(strconv.Itoa(c.index) + c.inputs[c.index])
				proof, err := tree.ProofFor(indexedHash)
				if err != nil {
					t.Fatal(err)
				}

				for i := 0; i < len(c.expectedProof); i++ {
					if got, want := string(proof.hashes[i]), c.expectedProof[i]; got != want {
						t.Errorf("got %v, want %v", got, want)
					}
				}

				if err := proof.Verify(indexedHash, tree.root); err != nil {
					t.Errorf(err.Error())
				}
			})
		}
	})

	t.Run("From", func(t *testing.T) {
		tree, err := From(stringsToBytes([]string{"a", "b"}))
		if err != nil {
			t.Fatal(err)
		}
		left := tree.newHash().Sum([]byte("a"))
		right := tree.newHash().Sum([]byte("b"))

		if got, want := tree.root, tree.newHash().Sum(append(left, right...)); reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		if got, want := tree.height, 2; got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("FromHashes", func(t *testing.T) {
		a := sha256.New().Sum([]byte("a"))
		b := sha256.New().Sum([]byte("b"))

		tree, err := FromHashes([][]byte{a, b}, sha256.New)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := tree.root, tree.newHash().Sum(append(a, b...)); reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
		if got, want := tree.height, 2; got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("new invalid inputs", func(t *testing.T) {
		_, err := From(nil)
		if got, want := err.Error(), "invalid inputs"; !strings.Contains(got, want) {
			t.Fatalf("'%v', does not contains %v", got, want)
		}
	})
}

func stringsToBytes(in []string) [][]byte {
	var out [][]byte
	for _, i := range in {
		out = append(out, []byte(i))
	}
	return out
}
