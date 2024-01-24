package merklestore

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/tclairet/merklestore/merkletree"
)

const rootFileName = "root"

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
	if err := u.fileHandler.save(rootFileName, []byte(hex.EncodeToString(root))); err != nil {
		return err
	}
	return nil
}

func (u Uploader) Download(name string) error {
	root, err := u.getRoot()
	if err != nil {
		return err
	}
	file, proof, err := u.server.Request(name)
	if err != nil {
		return err
	}
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return err
	}
	return proof.Verify(hasher.Sum(nil), root)
}

func (u Uploader) upload(path string) error {
	file, err := u.fileHandler.open(path)
	if err != nil {
		return err
	}
	var r2 bytes.Buffer
	r1 := io.TeeReader(file, &r2)

	defer file.Close()
	if err := u.server.Upload(filepath.Base(path), r1); err != nil {
		return err
	}
	if err := u.builder.Add(&r2); err != nil {
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

func (u Uploader) getRoot() ([]byte, error) {
	buf, err := u.fileHandler.open(rootFileName)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(buf)
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(string(b))
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
