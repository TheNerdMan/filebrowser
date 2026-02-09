package scanner

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// ScanStatus represents the security scan status of a file
type ScanStatus string

const (
	StatusUploading    ScanStatus = "uploading"
	StatusScanning     ScanStatus = "scanning"
	StatusClean        ScanStatus = "clean"
	StatusSecurityRisk ScanStatus = "security_risk"
	StatusScanError    ScanStatus = "scan_error"
	StatusOverridden   ScanStatus = "overridden" // Admin marked as false positive
)

// ScanResult represents the result of a virus scan
type ScanResult struct {
	Status    ScanStatus
	Signature string
	Error     error
}

// Scanner interface for virus scanning
type Scanner interface {
	ScanFile(ctx context.Context, filePath string) (*ScanResult, error)
	IsAvailable() bool
}

// ClamAVScanner implements Scanner using ClamAV
type ClamAVScanner struct {
	host    string
	port    string
	timeout time.Duration
}

// NewClamAVScanner creates a new ClamAV scanner
func NewClamAVScanner(host, port string) *ClamAVScanner {
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "3310"
	}
	return &ClamAVScanner{
		host:    host,
		port:    port,
		timeout: 30 * time.Second,
	}
}

// IsAvailable checks if ClamAV is available
func (s *ClamAVScanner) IsAvailable() bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", s.host, s.port), 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Send PING command
	_, err = conn.Write([]byte("zPING\x00"))
	if err != nil {
		return false
	}

	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		return false
	}

	response := string(buf[:n])
	return strings.Contains(response, "PONG")
}

// ScanFile scans a file for viruses using ClamAV
func (s *ClamAVScanner) ScanFile(ctx context.Context, filePath string) (*ScanResult, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", s.host, s.port), s.timeout)
	if err != nil {
		return &ScanResult{
			Status: StatusScanError,
			Error:  fmt.Errorf("failed to connect to ClamAV: %w", err),
		}, nil
	}
	defer conn.Close()

	// Set connection deadline
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(s.timeout)
	}
	conn.SetDeadline(deadline)

	// Send scan command using SCAN protocol
	command := fmt.Sprintf("zSCAN %s\x00", filePath)
	_, err = conn.Write([]byte(command))
	if err != nil {
		return &ScanResult{
			Status: StatusScanError,
			Error:  fmt.Errorf("failed to send scan command: %w", err),
		}, nil
	}

	// Read response
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return &ScanResult{
			Status: StatusScanError,
			Error:  fmt.Errorf("failed to read scan result: %w", err),
		}, nil
	}

	response := string(buf[:n])
	response = strings.TrimSpace(response)

	// Parse ClamAV response
	// Response format: "/path/to/file: virusname FOUND" or "/path/to/file: OK"
	if strings.HasSuffix(response, "OK") {
		return &ScanResult{
			Status: StatusClean,
		}, nil
	}

	if strings.Contains(response, "FOUND") {
		parts := strings.Split(response, ":")
		if len(parts) >= 2 {
			virusInfo := strings.TrimSpace(parts[1])
			virusInfo = strings.TrimSuffix(virusInfo, " FOUND")
			return &ScanResult{
				Status:    StatusSecurityRisk,
				Signature: virusInfo,
			}, nil
		}
	}

	// If we can't parse the response, treat it as an error
	return &ScanResult{
		Status: StatusScanError,
		Error:  fmt.Errorf("unexpected ClamAV response: %s", response),
	}, nil
}

// NoOpScanner is a scanner that does nothing (for when ClamAV is not available)
type NoOpScanner struct{}

// NewNoOpScanner creates a new no-op scanner
func NewNoOpScanner() *NoOpScanner {
	return &NoOpScanner{}
}

// IsAvailable always returns false for NoOpScanner
func (s *NoOpScanner) IsAvailable() bool {
	return false
}

// ScanFile always returns clean status for NoOpScanner
func (s *NoOpScanner) ScanFile(ctx context.Context, filePath string) (*ScanResult, error) {
	return &ScanResult{
		Status: StatusClean,
	}, nil
}
