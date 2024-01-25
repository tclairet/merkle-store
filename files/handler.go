package files

import (
	"io"
	"os"
	"path/filepath"
)

type Handler interface {
	Open(path string) (io.ReadCloser, error)
	Delete(path string) error
	Save(name string, content io.Reader) error
}

type OS struct{}

func (OS) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (OS) Delete(path string) error {
	return os.RemoveAll(path)
}

func (OS) Save(name string, content io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	dst, err := os.Create(name)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, content); err != nil {
		return err
	}
	return nil
}
