package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/tclairet/merklestore/files"
	"github.com/tclairet/merklestore/merkletree"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

type Server struct {
	files    files.Handler
	db       store
	builders map[string]*merkletree.IndexedBuilder
	trees    map[string]*merkletree.MerkleTree
}

func New(files files.Handler, db store) (*Server, error) {
	savedTrees := db.read()
	trees := make(map[string]*merkletree.MerkleTree)
	for root, hashes := range savedTrees {
		tree, err := merkletree.FromHashes(hashes)
		if err != nil {
			return nil, err
		}
		trees[root] = tree
	}
	return &Server{
		files:    files,
		db:       db,
		builders: make(map[string]*merkletree.IndexedBuilder),
		trees:    trees,
	}, nil
}

func (s *Server) Upload(root string, index, total int, file io.Reader) error {
	if err := s.files.Save(fmt.Sprintf("%s/%d", root, index), file); err != nil {
		return err
	}
	reader, err := s.files.Open(fmt.Sprintf("%s/%d", root, index))
	if err != nil {
		return err
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return err
	}

	if err := s.db.save(root, hasher.Sum(nil), index, total); err != nil {
		return err
	}

	if s.builders[root] == nil {
		s.builders[root] = merkletree.NewIndexedBuilder(total)
	}

	done, err := s.builders[root].AddHash(index, hasher.Sum(nil))
	if err != nil {
		return err
	}

	logger.Info("uploaded",
		"root", root,
		"index", index,
		"hash", hex.EncodeToString(hasher.Sum(nil)),
	)

	if !done {
		return nil
	}

	s.trees[root], err = s.builders[root].Build()
	if err != nil {
		return err
	}

	delete(s.builders, root)
	return nil
}

func (s *Server) Request(root string, index int) (io.Reader, *merkletree.Proof, error) {
	if s.trees[root] == nil {
		return nil, nil, fmt.Errorf("unknow or unfinished tree")
	}
	file, err := s.files.Open(fmt.Sprintf("%s/%d", root, index))
	if err != nil {
		return nil, nil, err
	}
	hash, err := s.db.get(root, index)
	if err != nil {
		return nil, nil, err
	}
	proof, err := s.trees[root].ProofFor(hash)
	if err != nil {
		return nil, nil, err
	}
	logger.Info("request",
		"root", root,
		"index", index,
		"proof", func() (hashes []string) {
			for _, h := range proof.Hashes() {
				hashes = append(hashes, hex.EncodeToString(h))
			}
			return
		}(),
	)
	return file, proof, nil
}
