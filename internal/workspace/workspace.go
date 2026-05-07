package workspace

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrInvalidPath = errors.New("invalid project path")

type Workspace struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type Service struct {
	current *Workspace
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Open(path string) (*Workspace, error) {
	if path == "" {
		return nil, ErrInvalidPath
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, ErrInvalidPath
	}
	ws := &Workspace{Path: abs, Name: filepath.Base(abs)}
	s.current = ws
	return ws, nil
}

func (s *Service) Current() *Workspace {
	return s.current
}

func (s *Service) Close() {
	s.current = nil
}
