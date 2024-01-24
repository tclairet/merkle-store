package merklestore

import (
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/tclairet/merklestore/merkletree"
)

type Uploader struct {
	server  Server
	builder *merkletree.Builder

	fileHandler fileHandler
}

func NewUploader() *Uploader {
	return &Uploader{
		server:      nil,
		builder:     merkletree.NewBuilder(),
		fileHandler: osWrapper{},
	}
}

func (u Uploader) Upload(paths []string) error {
	for _, path := range paths {
		if err := u.upload(path); err != nil {
			return err
		}
		if err := u.delete(path); err != nil {
			return err
		}
	}
	tree, err := u.builder.Build()
	if err != nil {
		return err
	}
	root := tree.Root()
	if err := u.fileHandler.save("root", []byte(hex.EncodeToString(root))); err != nil {
		return err
	}
	return nil
}

func (u Uploader) upload(path string) error {
	file, err := u.fileHandler.open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := u.server.Upload(filepath.Base(path), file); err != nil {
		return err
	}
	if err := u.builder.Add(file); err != nil {
		return err
	}
	return nil
}

func (u Uploader) delete(path string) error {
	if err := u.fileHandler.delete(path); err != nil {
		return err
	}
	return nil
}

type fileHandler interface {
	open(path string) (io.ReadCloser, error)
	delete(path string) error
	save(name string, content []byte) error
}

type osWrapper struct{}

func (osWrapper) open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (osWrapper) delete(path string) error {
	return os.RemoveAll(path)
}

func (osWrapper) save(name string, content []byte) error {
	return os.WriteFile(name, content, 0666)
}
