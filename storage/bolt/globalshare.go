package bolt

import (
	"errors"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/share"
)

type globalShareRequestBackend struct {
	db *storm.DB
}

func (s globalShareRequestBackend) GetRequest(id uint) (*share.GlobalShareRequest, error) {
	var v share.GlobalShareRequest
	err := s.db.One("ID", id, &v)
	if errors.Is(err, storm.ErrNotFound) {
		return nil, fberrors.ErrNotExist
	}
	return &v, err
}

func (s globalShareRequestBackend) GetAllRequests() ([]*share.GlobalShareRequest, error) {
	var v []*share.GlobalShareRequest
	err := s.db.All(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return v, nil
	}
	return v, err
}

func (s globalShareRequestBackend) GetRequestsByStatus(status string) ([]*share.GlobalShareRequest, error) {
	var v []*share.GlobalShareRequest
	err := s.db.Select(q.Eq("Status", status)).Find(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return v, nil
	}
	return v, err
}

func (s globalShareRequestBackend) GetRequestsByUserID(userID uint) ([]*share.GlobalShareRequest, error) {
	var v []*share.GlobalShareRequest
	err := s.db.Select(q.Eq("UserID", userID)).Find(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return v, nil
	}
	return v, err
}

func (s globalShareRequestBackend) SaveRequest(req *share.GlobalShareRequest) error {
	return s.db.Save(req)
}

func (s globalShareRequestBackend) DeleteRequest(id uint) error {
	err := s.db.DeleteStruct(&share.GlobalShareRequest{ID: id})
	if errors.Is(err, storm.ErrNotFound) {
		return nil
	}
	return err
}

type globalShareBackend struct {
	db *storm.DB
}

func (s globalShareBackend) GetGlobalShare(id uint) (*share.GlobalShare, error) {
	var v share.GlobalShare
	err := s.db.One("ID", id, &v)
	if errors.Is(err, storm.ErrNotFound) {
		return nil, fberrors.ErrNotExist
	}
	return &v, err
}

func (s globalShareBackend) GetAllGlobalShares() ([]*share.GlobalShare, error) {
	var v []*share.GlobalShare
	err := s.db.All(&v)
	if errors.Is(err, storm.ErrNotFound) {
		return v, nil
	}
	return v, err
}

func (s globalShareBackend) GetGlobalShareByPath(path string) (*share.GlobalShare, error) {
	var v share.GlobalShare
	err := s.db.One("Path", path, &v)
	if errors.Is(err, storm.ErrNotFound) {
		return nil, fberrors.ErrNotExist
	}
	return &v, err
}

func (s globalShareBackend) SaveGlobalShare(gs *share.GlobalShare) error {
	return s.db.Save(gs)
}

func (s globalShareBackend) DeleteGlobalShare(id uint) error {
	err := s.db.DeleteStruct(&share.GlobalShare{ID: id})
	if errors.Is(err, storm.ErrNotFound) {
		return nil
	}
	return err
}
