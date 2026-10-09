package file

import (
	"context"
	"fmt"
	"sync"
)

// StubStorage is a test double for file storage.
// It stores files in memory and returns a fake URL.
type StubStorage struct {
	mu    sync.Mutex
	Files map[string][]byte
}

func NewStub() *StubStorage {
	return &StubStorage{
		Files: make(map[string][]byte),
	}
}

// GetFile returns the content of a stored file by path.
func (s *StubStorage) GetFile(path string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, ok := s.Files[path]
	if !ok {
		return nil, false
	}

	result := make([]byte, len(content))
	copy(result, content)
	return result, true
}

func (s *StubStorage) StoreFile(_ context.Context, path string, content []byte) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored := make([]byte, len(content))
	copy(stored, content)
	s.Files[path] = stored

	return fmt.Sprintf("https://stub-storage.local/%s", path), nil
}
