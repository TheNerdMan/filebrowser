package fbhttp

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/filebrowser/filebrowser/v2/scanner"
)

// Re-export scanner types for convenience
type (
	ClamAVScanner    = scanner.ClamAVScanner
	MemoryScanStore  = scanner.MemoryStore
	FileScanInfo     = scanner.FileScanInfo
	ScanStatus       = scanner.ScanStatus
)

// Constructor aliases
var (
	NewClamAVScanner   = scanner.NewClamAVScanner
	NewMemoryScanStore = scanner.NewMemoryStore
)

// ScannerService defines the interface for the scanner service
type ScannerService interface {
	ScanFile(ctx context.Context, filePath string, userID uint) error
	GetScanStatus(path string) (*scanner.FileScanInfo, error)
	ListSecurityRisks() ([]*scanner.FileScanInfo, error)
	DeleteScanInfo(path string) error
	IsAvailable() bool
}

// DefaultScannerService implements ScannerService
type DefaultScannerService struct {
	scanner scanner.Scanner
	store   scanner.Store
}

// NewScannerService creates a new scanner service
func NewScannerService(s scanner.Scanner, store scanner.Store) *DefaultScannerService {
	return &DefaultScannerService{
		scanner: s,
		store:   store,
	}
}

// ScanFile scans a file and updates its status
func (s *DefaultScannerService) ScanFile(ctx context.Context, filePath string, userID uint) error {
	// Mark as scanning
	info := &scanner.FileScanInfo{
		Path:       filePath,
		UserID:     userID,
		Status:     scanner.StatusScanning,
		UploadedAt: time.Now(),
	}
	if err := s.store.Set(info); err != nil {
		return err
	}

	// Perform scan
	result, err := s.scanner.ScanFile(ctx, filePath)
	if err != nil {
		info.Status = scanner.StatusScanError
		s.store.Set(info)
		return err
	}

	// Update status based on result
	info.Status = result.Status
	info.Signature = result.Signature
	info.ScannedAt = time.Now()
	return s.store.Set(info)
}

// GetScanStatus returns the scan status for a file
func (s *DefaultScannerService) GetScanStatus(path string) (*scanner.FileScanInfo, error) {
	return s.store.Get(path)
}

// ListSecurityRisks returns all files with security risks
func (s *DefaultScannerService) ListSecurityRisks() ([]*scanner.FileScanInfo, error) {
	return s.store.ListByStatus(scanner.StatusSecurityRisk)
}

// DeleteScanInfo deletes scan information for a file
func (s *DefaultScannerService) DeleteScanInfo(path string) error {
	return s.store.Delete(path)
}

// IsAvailable checks if scanner is available
func (s *DefaultScannerService) IsAvailable() bool {
	return s.scanner.IsAvailable()
}

// scanStatusHandler returns the scan status for a file
func scanStatusHandler(scannerSvc ScannerService) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		path := r.URL.Query().Get("path")
		if path == "" {
			return http.StatusBadRequest, nil
		}

		// Resolve to full path
		fullPath := filepath.Join(d.settings.Defaults.Scope, path)
		if d.user != nil {
			fullPath = d.user.FullPath(path)
		}
		
		info, err := scannerSvc.GetScanStatus(fullPath)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if info == nil {
			// No scan info means the file hasn't been scanned yet
			return renderJSON(w, r, map[string]string{"status": string(scanner.StatusClean)})
		}

		return renderJSON(w, r, info)
	})
}

// scanRisksHandler returns all files with security risks (admin only)
func scanRisksHandler(scannerSvc ScannerService) handleFunc {
	return withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		risks, err := scannerSvc.ListSecurityRisks()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if risks == nil {
			risks = []*scanner.FileScanInfo{}
		}

		return renderJSON(w, r, risks)
	})
}

// scanRiskDeleteHandler deletes a file with security risk (admin only)
func scanRiskDeleteHandler(scannerSvc ScannerService) handleFunc {
	return withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		path := r.URL.Query().Get("path")
		if path == "" {
			return http.StatusBadRequest, nil
		}

		// Get scan info to verify it's a security risk
		info, err := scannerSvc.GetScanStatus(path)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if info == nil || info.Status != scanner.StatusSecurityRisk {
			return http.StatusNotFound, nil
		}

		// Delete the actual file
		// We need to find which user owns this file and use their filesystem
		// For now, we'll use the admin's filesystem since this is an admin-only endpoint
		// Note: This is a simplified approach - the path is expected to be the full system path
		_, err = d.store.Settings.Get()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		// Try to remove the file from the filesystem
		// Note: This is a simplified approach - in production you'd want to find the actual user's Fs
		// For now we assume the path is already the full path
		err = d.user.Fs.Remove(path)
		if err != nil {
			log.Printf("Failed to delete file %s: %v", path, err)
			return http.StatusInternalServerError, err
		}

		// Delete scan info
		err = scannerSvc.DeleteScanInfo(path)
		if err != nil {
			log.Printf("Failed to delete scan info for %s: %v", path, err)
		}

		return http.StatusNoContent, nil
	})
}
