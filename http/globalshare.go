package fbhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/share"
)

// Request Global Share - any user can request
var globalShareRequestHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	var body struct {
		Path    string `json:"path"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return http.StatusBadRequest, err
	}
	defer r.Body.Close()

	if body.Path == "" {
		return http.StatusBadRequest, errors.New("path is required")
	}

	// Create the request
	req := &share.GlobalShareRequest{
		Path:      body.Path,
		UserID:    d.user.ID,
		Username:  d.user.Username,
		Status:    "pending",
		CreatedAt: time.Now().Unix(),
		Message:   body.Message,
	}

	if err := d.store.GlobalShareRequest.Save(req); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, req)
})

// List all global share requests - only for users with ManageGlobalShare permission
var globalShareRequestListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.ManageGlobalShare && !d.user.Perm.Admin {
		return http.StatusForbidden, nil
	}

	requests, err := d.store.GlobalShareRequest.GetAll()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, requests)
})

// Get user's own requests
var globalShareMyRequestsHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	requests, err := d.store.GlobalShareRequest.GetByUserID(d.user.ID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, requests)
})

// Approve/Reject a global share request - only for users with ManageGlobalShare permission
var globalShareRequestActionHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.ManageGlobalShare && !d.user.Perm.Admin {
		return http.StatusForbidden, nil
	}

	// Get request ID from URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/globalshare/requests/")
	parts := strings.Split(idStr, "/")
	if len(parts) < 2 {
		return http.StatusBadRequest, errors.New("invalid URL format")
	}

	id, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return http.StatusBadRequest, err
	}

	action := parts[1] // "approve" or "reject"

	req, err := d.store.GlobalShareRequest.Get(uint(id))
	if err != nil {
		if errors.Is(err, fberrors.ErrNotExist) {
			return http.StatusNotFound, err
		}
		return http.StatusInternalServerError, err
	}

	if action == "approve" {
		// Create a global share entry
		globalShare := &share.GlobalShare{
			Path:         req.Path,
			OriginalPath: req.Path,
			UserID:       req.UserID,
			Username:     req.Username,
			AddedAt:      time.Now().Unix(),
			RequestID:    req.ID,
		}

		if err := d.store.GlobalShare.Save(globalShare); err != nil {
			return http.StatusInternalServerError, err
		}

		// Update request status
		req.Status = "approved"
	} else if action == "reject" {
		req.Status = "rejected"
	} else {
		return http.StatusBadRequest, errors.New("invalid action")
	}

	if err := d.store.GlobalShareRequest.Save(req); err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, req)
})

// Delete a global share request
var globalShareRequestDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/globalshare/requests/")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return http.StatusBadRequest, err
	}

	req, err := d.store.GlobalShareRequest.Get(uint(id))
	if err != nil {
		if errors.Is(err, fberrors.ErrNotExist) {
			return http.StatusNotFound, err
		}
		return http.StatusInternalServerError, err
	}

	// Only allow deletion by the requester or admin/manager
	if req.UserID != d.user.ID && !d.user.Perm.Admin && !d.user.Perm.ManageGlobalShare {
		return http.StatusForbidden, nil
	}

	if err := d.store.GlobalShareRequest.Delete(uint(id)); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
})

// List all global shares - accessible to all users
var globalShareListHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	shares, err := d.store.GlobalShare.GetAll()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return renderJSON(w, r, shares)
})

// Delete a global share - only for users with ManageGlobalShare permission
var globalShareDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.ManageGlobalShare && !d.user.Perm.Admin {
		return http.StatusForbidden, nil
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/api/globalshare/")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if err := d.store.GlobalShare.Delete(uint(id)); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
})
