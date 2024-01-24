package merkletree

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"reflect"
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
		tree := MerkleTree{newHash: newFakeHash}

		cases := []struct {
			inputs        []string
			hash          string
			expectedProof []string
		}{
			{inputs: []string{"a"}, hash: "a", expectedProof: []string{"a"}},
			{inputs: []string{"a", "b"}, hash: "a", expectedProof: []string{"a", "b", "ab"}},
			{inputs: []string{"a", "b", "c"}, hash: "a", expectedProof: []string{"a", "b", "ab", "c", "abc"}},
			{inputs: []string{"a", "b", "c"}, hash: "c", expectedProof: []string{"c", "ab", "abc"}},
			{inputs: []string{"a", "b", "c", "d"}, hash: "a", expectedProof: []string{"a", "b", "ab", "cd", "abcd"}},
			{inputs: []string{"a", "b", "c", "d"}, hash: "b", expectedProof: []string{"b", "a", "ab", "cd", "abcd"}},
			{inputs: []string{"a", "b", "c", "d"}, hash: "c", expectedProof: []string{"c", "d", "cd", "ab", "abcd"}},
			{inputs: []string{"a", "b", "c", "d"}, hash: "d", expectedProof: []string{"d", "c", "cd", "ab", "abcd"}},
			{inputs: []string{"a", "b", "c", "d", "e"}, hash: "a", expectedProof: []string{"a", "b", "ab", "cd", "abcd", "e", "abcde"}},
			{inputs: []string{"a", "b", "c", "d", "e"}, hash: "e", expectedProof: []string{"e", "abcd", "abcde"}},
			{inputs: []string{"a", "b", "c", "d", "e", "f"}, hash: "a", expectedProof: []string{"a", "b", "ab", "cd", "abcd", "ef", "abcdef"}},
			{inputs: []string{"a", "b", "c", "d", "e", "f", "g"}, hash: "a", expectedProof: []string{"a", "b", "ab", "cd", "abcd", "efg", "abcdefg"}},
			{inputs: []string{"a", "b", "c", "d", "e", "f", "g", "h"}, hash: "a", expectedProof: []string{"a", "b", "ab", "cd", "abcd", "efgh", "abcdefgh"}},
			{inputs: []string{"a", "b", "c", "d", "e", "f", "g", "h"}, hash: "d", expectedProof: []string{"d", "c", "cd", "ab", "abcd", "efgh", "abcdefgh"}},
			{inputs: []string{"a", "b", "c", "d", "e", "f", "g", "h"}, hash: "h", expectedProof: []string{"h", "g", "gh", "ef", "efgh", "abcd", "abcdefgh"}},
		}

		for _, c := range cases {
			t.Run(fmt.Sprintf("%s for %s", strings.Join(c.inputs, ""), c.hash), func(t *testing.T) {
				tree.from(stringsToBytes(c.inputs))

				proof, err := tree.ProofFor([]byte(c.hash))
				if err != nil {
					t.Fatal(err)
				}

				for i := 0; i < len(c.expectedProof); i++ {
					if got, want := string(proof.hashes[i]), c.expectedProof[i]; got != want {
						t.Errorf("got %v, want %v", got, want)
					}
				}

				if err := proof.Verify([]byte(c.hash), tree.root); err != nil {
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

		tree, err := FromHashes([][]byte{a, b})
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
