package filesystem

import (
	"context"
	"fmt"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/ports"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Storage struct{ root string }

func New(root string) (*Storage, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err = os.MkdirAll(root, 0750); err != nil {
		return nil, err
	}
	return &Storage{root: root}, nil
}
func (s *Storage) path(key string) (string, error) {
	value := filepath.Join(s.root, filepath.FromSlash(key))
	if !strings.HasPrefix(value, s.root+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid object key")
	}
	return value, nil
}
func (s *Storage) Put(_ context.Context, key string, body io.Reader, _ int64, _ string) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "upload-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err = io.Copy(tmp, body); err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(name, path)
}
func (s *Storage) Open(_ context.Context, key string) (io.ReadCloser, error) {
	path, err := s.path(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}
func (s *Storage) Stat(_ context.Context, key string) (ports.ObjectInfo, error) {
	path, err := s.path(key)
	if err != nil {
		return ports.ObjectInfo{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return ports.ObjectInfo{}, err
	}
	return ports.ObjectInfo{Size: info.Size()}, nil
}
func (s *Storage) PresignPut(context.Context, string, string, time.Duration) (string, error) {
	return "", nil
}
