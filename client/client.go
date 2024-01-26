package client

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"syscall"

	"github.com/tclairet/merklestore/files"
	"github.com/tclairet/merklestore/merkletree"
	"github.com/tclairet/merklestore/server"
)

const rootFileName = "root"

var _ Server = server.Client{}

type Server interface {
	Upload(root string, index, total int, file io.Reader) error
	Request(root string, index int) (io.Reader, *merkletree.Proof, error)
}

type Uploader struct {
	server  Server
	builder *merkletree.Builder

	fileHandler files.Handler
}

func NewUploader(handler files.Handler, server Server) *Uploader {
	return &Uploader{
		server:      server,
		builder:     merkletree.NewBuilder(),
		fileHandler: handler,
	}
}

func (u Uploader) Upload(paths []string) (string, error) {
	root, err := u.root(paths)
	if err != nil {
		return "", err
	}
	for i, path := range paths {
		if err := u.upload(root, path, i, len(paths)); err != nil {
			return "", err
		}
		if err := u.delete(path); err != nil {
			return "", err
		}
	}
	return root, nil
}

func (u Uploader) root(paths []string) (string, error) {
	for _, path := range paths {
		file, err := u.fileHandler.Open(path)
		if err != nil {
			return "", err
		}
		if err := u.builder.Add(file); err != nil {
			return "", err
		}
	}
	tree, err := u.builder.Build()
	if err != nil {
		return "", err
	}

	root := hex.EncodeToString(tree.Root())
	if err := u.saveRoot(root); err != nil {
		return "", err
	}
	return root, nil
}

func (u Uploader) Download(root string, indexes ...int) error {
	roots, err := u.getRoots()
	if err != nil {
		return err
	}
	if !slices.Contains(roots, root) {
		return fmt.Errorf("unknown root hash")
	}
	for _, index := range indexes {
		if err := u.downloadIndex(root, index); err != nil {
			return err
		}
	}
	return nil
}

func (u Uploader) downloadIndex(root string, index int) error {
	file, proof, err := u.server.Request(root, index)
	if err != nil {
		return err
	}

	if err := u.fileHandler.Save(fmt.Sprintf("%s/%d", root, index), file); err != nil {
		return err
	}

	reader, err := u.fileHandler.Open(fmt.Sprintf("%s/%d", root, index))
	if err != nil {
		return err
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return err
	}
	b, err := hex.DecodeString(root)
	if err != nil {
		return err
	}
	return proof.Verify(hasher.Sum(nil), b)
}

func (u Uploader) upload(root, path string, i, total int) error {
	file, err := u.fileHandler.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()
	if err := u.server.Upload(root, i, total, file); err != nil {
		return err
	}
	return nil
}

func (u Uploader) delete(path string) error {
	if err := u.fileHandler.Delete(path); err != nil {
		return err
	}
	return nil
}

type RootsBackup struct {
	Roots []string `json:"roots"`
}

func (u Uploader) saveRoot(root string) error {
	roots, err := u.getRoots()
	if err != nil {
		return err
	}
	b, err := json.Marshal(RootsBackup{Roots: append(roots, root)})
	if err != nil {
		return err
	}
	if err := u.fileHandler.Save(rootFileName, bytes.NewBuffer(b)); err != nil {
		return err
	}
	return nil
}

func (u Uploader) getRoots() ([]string, error) {
	f, err := u.fileHandler.Open(rootFileName)
	if err != nil {
		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			if errors.Is(pathErr.Err, syscall.ENOENT) {
				return []string{}, nil
			}
		}
		return nil, err
	}
	var rootsBackup RootsBackup
	if err := json.NewDecoder(f).Decode(&rootsBackup); err != nil {
		return nil, err
	}
	return rootsBackup.Roots, nil
}
