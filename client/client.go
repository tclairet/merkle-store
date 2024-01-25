package client

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

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

func (u Uploader) Upload(paths []string) error {
	root, err := u.root(paths)
	if err != nil {
		return err
	}
	for i, path := range paths {
		if err := u.upload(root, path, i, len(paths)); err != nil {
			return err
		}
		if err := u.delete(path); err != nil {
			return err
		}
	}
	return nil
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
	root := tree.Root()
	if err := u.fileHandler.Save(rootFileName, bytes.NewBuffer([]byte(hex.EncodeToString(root)))); err != nil {
		return "", err
	}
	return hex.EncodeToString(root), nil
}

func (u Uploader) Download(index int) error {
	root, err := u.getRoot()
	if err != nil {
		return err
	}
	file, proof, err := u.server.Request(hex.EncodeToString(root), index)
	if err != nil {
		return err
	}

	if err := u.fileHandler.Save(fmt.Sprintf("%d", index), file); err != nil {
		return err
	}

	reader, err := u.fileHandler.Open(fmt.Sprintf("%d", index))
	if err != nil {
		return err
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return err
	}

	return proof.Verify(hasher.Sum(nil), root)
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

func (u Uploader) getRoot() ([]byte, error) {
	buf, err := u.fileHandler.Open(rootFileName)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(buf)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(string(b))
}
