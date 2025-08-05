package internal

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

var fileSys realFileSystem

type (
	fileSystem interface {
		Close(closer io.Closer) error
		Copy(writer io.Writer, reader io.Reader) (int64, error)
		Create(name string) (*os.File, error)
		Open(name string) (*os.File, error)
		Remove(name string) error
	}

	realFileSystem struct{}
)

func (fs realFileSystem) Close(closer io.Closer) error {
	return closer.Close()
}

func (fs realFileSystem) Copy(writer io.Writer, reader io.Reader) (int64, error) {
	return io.Copy(writer, reader)
}

func validatePath(name string) error {
	clean := filepath.Clean(name)
	if strings.Contains(clean, "..") {
		return os.ErrPermission
	}
	return nil
}

func (fs realFileSystem) Create(name string) (*os.File, error) {
	if err := validatePath(name); err != nil {
		return nil, err
	}
	return os.Create(filepath.Clean(name))
}

func (fs realFileSystem) Open(name string) (*os.File, error) {
	if err := validatePath(name); err != nil {
		return nil, err
	}
	return os.Open(filepath.Clean(name))
}

func (fs realFileSystem) Remove(name string) error {
	if err := validatePath(name); err != nil {
		return err
	}
	return os.Remove(filepath.Clean(name))
}
