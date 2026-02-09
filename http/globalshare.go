package fbhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/fileutils"
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
		// Get the user who made the request to access their filesystem
		requestUser, err := d.store.Users.Get(d.server.Root, req.UserID)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to get request user: %w", err)
		}

		// Define the global share directory path (relative to server root)
		globalShareDir := filepath.Join(d.server.Root, ".globalshare")
		
		// Create global share directory if it doesn't exist with restrictive permissions
		if err := os.MkdirAll(globalShareDir, 0750); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to create global share directory: %w", err)
		}

		// Get the source file's real path
		sourcePath := requestUser.FullPath(req.Path)
		
		// Create a unique destination filename to avoid conflicts
		timestamp := time.Now().Unix()
		baseName := filepath.Base(req.Path)
		destFileName := fmt.Sprintf("%d_%s_%s", timestamp, req.Username, baseName)
		destPath := filepath.Join(globalShareDir, destFileName)

		// Copy the file/folder from the user's scope to the global share directory
		// Use the OS filesystem for this operation since we're working with real paths
		// Use restrictive permissions (read-only) for security
		osFs := afero.NewOsFs()
		if err := fileutils.Copy(osFs, sourcePath, destPath, 0440, 0750); err != nil {
			return http.StatusInternalServerError, fmt.Errorf("failed to copy file to global share: %w", err)
		}

		// Create a global share entry with the new path
		globalShare := &share.GlobalShare{
			Path:         filepath.Join("/.globalshare", destFileName), // Virtual path for users
			OriginalPath: req.Path,
			UserID:       req.UserID,
			Username:     req.Username,
			AddedAt:      time.Now().Unix(),
			RequestID:    req.ID,
		}

		if err := d.store.GlobalShare.Save(globalShare); err != nil {
			// Cleanup the copied file if database save fails
			if removeErr := os.RemoveAll(destPath); removeErr != nil {
				// Log the cleanup failure but return the original error
				log.Printf("Failed to cleanup copied file %s after database error: %v", destPath, removeErr)
			}
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

// Delete a global share - only for admins
var globalShareDeleteHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Admin {
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
