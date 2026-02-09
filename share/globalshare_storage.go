package share

import (
	fberrors "github.com/filebrowser/filebrowser/v2/errors"
)

// GlobalShareRequestBackend is the interface for global share request storage.
type GlobalShareRequestBackend interface {
	GetRequest(id uint) (*GlobalShareRequest, error)
	GetAllRequests() ([]*GlobalShareRequest, error)
	GetRequestsByStatus(status string) ([]*GlobalShareRequest, error)
	GetRequestsByUserID(userID uint) ([]*GlobalShareRequest, error)
	SaveRequest(req *GlobalShareRequest) error
	DeleteRequest(id uint) error
}

// GlobalShareBackend is the interface for global share storage.
type GlobalShareBackend interface {
	GetGlobalShare(id uint) (*GlobalShare, error)
	GetAllGlobalShares() ([]*GlobalShare, error)
	GetGlobalShareByPath(path string) (*GlobalShare, error)
	SaveGlobalShare(share *GlobalShare) error
	DeleteGlobalShare(id uint) error
}

// GlobalShareRequestStorage handles global share request operations.
type GlobalShareRequestStorage struct {
	back GlobalShareRequestBackend
}

// NewGlobalShareRequestStorage creates a new global share request storage.
func NewGlobalShareRequestStorage(back GlobalShareRequestBackend) *GlobalShareRequestStorage {
	return &GlobalShareRequestStorage{back: back}
}

// Get retrieves a global share request by ID.
func (s *GlobalShareRequestStorage) Get(id uint) (*GlobalShareRequest, error) {
	return s.back.GetRequest(id)
}

// GetAll retrieves all global share requests.
func (s *GlobalShareRequestStorage) GetAll() ([]*GlobalShareRequest, error) {
	return s.back.GetAllRequests()
}

// GetByStatus retrieves global share requests by status.
func (s *GlobalShareRequestStorage) GetByStatus(status string) ([]*GlobalShareRequest, error) {
	return s.back.GetRequestsByStatus(status)
}

// GetByUserID retrieves global share requests by user ID.
func (s *GlobalShareRequestStorage) GetByUserID(userID uint) ([]*GlobalShareRequest, error) {
	return s.back.GetRequestsByUserID(userID)
}

// Save saves a global share request.
func (s *GlobalShareRequestStorage) Save(req *GlobalShareRequest) error {
	return s.back.SaveRequest(req)
}

// Delete deletes a global share request.
func (s *GlobalShareRequestStorage) Delete(id uint) error {
	return s.back.DeleteRequest(id)
}

// GlobalShareStorage handles global share operations.
type GlobalShareStorage struct {
	back GlobalShareBackend
}

// NewGlobalShareStorage creates a new global share storage.
func NewGlobalShareStorage(back GlobalShareBackend) *GlobalShareStorage {
	return &GlobalShareStorage{back: back}
}

// Get retrieves a global share by ID.
func (s *GlobalShareStorage) Get(id uint) (*GlobalShare, error) {
	return s.back.GetGlobalShare(id)
}

// GetAll retrieves all global shares.
func (s *GlobalShareStorage) GetAll() ([]*GlobalShare, error) {
	return s.back.GetAllGlobalShares()
}

// GetByPath retrieves a global share by path.
func (s *GlobalShareStorage) GetByPath(path string) (*GlobalShare, error) {
	share, err := s.back.GetGlobalShareByPath(path)
	if err != nil {
		return nil, fberrors.ErrNotExist
	}
	return share, nil
}

// Save saves a global share.
func (s *GlobalShareStorage) Save(share *GlobalShare) error {
	return s.back.SaveGlobalShare(share)
}

// Delete deletes a global share.
func (s *GlobalShareStorage) Delete(id uint) error {
	return s.back.DeleteGlobalShare(id)
}
