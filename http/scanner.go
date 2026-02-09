package fbhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/filebrowser/filebrowser/v2/scanner"
	"github.com/filebrowser/filebrowser/v2/settings"
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
	OverrideSecurityRisk(path string, adminID uint) error
	QuarantineFile(path string, quarantinePath string) error
}

// DefaultScannerService implements ScannerService
type DefaultScannerService struct {
	scanner        scanner.Scanner
	store          scanner.Store
	settingsGetter func() (*settings.Settings, error)
}

// NewScannerService creates a new scanner service
func NewScannerService(s scanner.Scanner, store scanner.Store, settingsGetter func() (*settings.Settings, error)) *DefaultScannerService {
	return &DefaultScannerService{
		scanner:        s,
		store:          store,
		settingsGetter: settingsGetter,
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
	
	// If security risk detected, handle quarantine and webhook
	if result.Status == scanner.StatusSecurityRisk {
		// Get settings for webhook and quarantine
		if s.settingsGetter != nil {
			settings, err := s.settingsGetter()
			if err == nil && settings != nil {
				// Move to quarantine if configured
				if settings.QuarantinePath != "" {
					if err := s.QuarantineFile(filePath, settings.QuarantinePath); err != nil {
						log.Printf("Failed to quarantine file %s: %v", filePath, err)
					} else {
						log.Printf("File quarantined: %s", filePath)
					}
				}
				
				// Send webhook notification if configured
				if settings.SecurityWebhook != "" {
					go func() {
						if err := SendWebhookNotification(settings.SecurityWebhook, info); err != nil {
							log.Printf("Failed to send webhook notification: %v", err)
						}
					}()
				}
			}
		}
	}
	
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

// OverrideSecurityRisk marks a security risk as overridden by an admin
func (s *DefaultScannerService) OverrideSecurityRisk(path string, adminID uint) error {
	info, err := s.store.Get(path)
	if err != nil {
		return err
	}
	if info == nil {
		return fmt.Errorf("file not found in scan records")
	}
	
	info.Status = scanner.StatusOverridden
	info.OverriddenBy = adminID
	info.OverriddenAt = time.Now()
	return s.store.Set(info)
}

// QuarantineFile moves a file to quarantine
func (s *DefaultScannerService) QuarantineFile(path string, quarantinePath string) error {
	// Create quarantine directory if it doesn't exist
	if err := os.MkdirAll(quarantinePath, 0750); err != nil {
		return fmt.Errorf("failed to create quarantine directory: %w", err)
	}
	
	// Generate quarantine filename with timestamp
	filename := filepath.Base(path)
	timestamp := time.Now().Format("20060102-150405")
	quarantineFile := filepath.Join(quarantinePath, fmt.Sprintf("%s_%s", timestamp, filename))
	
	// Move the file
	if err := os.Rename(path, quarantineFile); err != nil {
		return fmt.Errorf("failed to move file to quarantine: %w", err)
	}
	
	// Update scan info with new path
	info, err := s.store.Get(path)
	if err == nil && info != nil {
		s.store.Delete(path)
		info.Path = quarantineFile
		s.store.Set(info)
	}
	
	return nil
}

// SendWebhookNotification sends a webhook notification for security events
func SendWebhookNotification(webhookURL string, info *scanner.FileScanInfo) error {
	if webhookURL == "" {
		return nil // No webhook configured
	}
	
	payload := map[string]interface{}{
		"event":     "security_risk_detected",
		"path":      info.Path,
		"userId":    info.UserID,
		"signature": info.Signature,
		"timestamp": info.ScannedAt.Format(time.RFC3339),
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}
	
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}
	
	return nil
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
		action := r.URL.Query().Get("action") // "delete" or "quarantine"
		
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

		if action == "quarantine" {
			// Move to quarantine instead of deleting
			quarantinePath := d.settings.QuarantinePath
			if quarantinePath == "" {
				quarantinePath = "/tmp/quarantine" // Default quarantine path
			}
			
			err = scannerSvc.QuarantineFile(path, quarantinePath)
			if err != nil {
				log.Printf("Failed to quarantine file %s: %v", path, err)
				return http.StatusInternalServerError, err
			}
			log.Printf("File quarantined by admin %d: %s", d.user.ID, path)
		} else {
			// Delete the actual file directly from the filesystem
			// The path stored in scan info is the full system path
			err = os.Remove(path)
			if err != nil && !os.IsNotExist(err) {
				log.Printf("Failed to delete file %s: %v", path, err)
				return http.StatusInternalServerError, err
			}

			// Delete scan info
			err = scannerSvc.DeleteScanInfo(path)
			if err != nil {
				log.Printf("Failed to delete scan info for %s: %v", path, err)
			}
		}

		return http.StatusNoContent, nil
	})
}

// scanRiskOverrideHandler marks a security risk as false positive (admin only)
func scanRiskOverrideHandler(scannerSvc ScannerService) handleFunc {
	return withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
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

		// Override the status
		err = scannerSvc.OverrideSecurityRisk(path, d.user.ID)
		if err != nil {
			log.Printf("Failed to override security risk for %s: %v", path, err)
			return http.StatusInternalServerError, err
		}

		log.Printf("Security risk overridden by admin %d: %s", d.user.ID, path)
		return http.StatusOK, nil
	})
}
