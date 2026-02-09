package scanner

import (
	"sync"
	"time"
)

// FileScanInfo stores scan metadata for a file
type FileScanInfo struct {
	Path       string     `json:"path"`
	UserID     uint       `json:"userId"`
	Status     ScanStatus `json:"status"`
	Signature  string     `json:"signature,omitempty"`
	ScannedAt  time.Time  `json:"scannedAt,omitempty"`
	UploadedAt time.Time  `json:"uploadedAt"`
}

// Store manages file scan information
type Store interface {
	Set(info *FileScanInfo) error
	Get(path string) (*FileScanInfo, error)
	Delete(path string) error
	ListByStatus(status ScanStatus) ([]*FileScanInfo, error)
	ListAll() ([]*FileScanInfo, error)
}

// MemoryStore implements Store using in-memory storage
type MemoryStore struct {
	mu    sync.RWMutex
	files map[string]*FileScanInfo
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		files: make(map[string]*FileScanInfo),
	}
}

// Set stores file scan information
func (s *MemoryStore) Set(info *FileScanInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[info.Path] = info
	return nil
}

// Get retrieves file scan information
func (s *MemoryStore) Get(path string) (*FileScanInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	info, ok := s.files[path]
	if !ok {
		return nil, nil
	}
	return info, nil
}

// Delete removes file scan information
func (s *MemoryStore) Delete(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.files, path)
	return nil
}

// ListByStatus returns all files with a specific status
func (s *MemoryStore) ListByStatus(status ScanStatus) ([]*FileScanInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var result []*FileScanInfo
	for _, info := range s.files {
		if info.Status == status {
			// Create a copy to prevent external modification
			infoCopy := *info
			result = append(result, &infoCopy)
		}
	}
	return result, nil
}

// ListAll returns all file scan information
func (s *MemoryStore) ListAll() ([]*FileScanInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*FileScanInfo, 0, len(s.files))
	for _, info := range s.files {
		// Create a copy to prevent external modification
		infoCopy := *info
		result = append(result, &infoCopy)
	}
	return result, nil
}
