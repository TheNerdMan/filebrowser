# ClamAV Security Scanning Feature

This implementation adds comprehensive virus scanning capabilities to File Browser using ClamAV.

## Features

### Backend
- **Real-time Virus Scanning**: All uploaded files are automatically scanned using ClamAV
- **Asynchronous Scanning**: Scans run in the background without blocking uploads
- **Graceful Degradation**: If ClamAV is unavailable, uploads continue normally without scanning
- **Download Protection**: Prevents downloading of files flagged as security risks
- **Admin Management**: Admins can view and delete all security risk files across all users

### Frontend
- **Upload Status Indicators**: Visual feedback showing "Uploading" → "Security Scanning" → "Uploaded"
- **Security Risk Alerts**: Files flagged by ClamAV display a clear warning
- **Admin Dashboard**: Dedicated view for managing security threats
- **Real-time Status Updates**: Automatic polling to show scan progress

## Configuration

### Environment Variables

The following environment variables can be set to configure ClamAV integration:

- `CLAMAV_HOST`: Hostname of the ClamAV daemon (default: "localhost")
- `CLAMAV_PORT`: Port of the ClamAV daemon (default: "3310")

### Settings (Admin Configuration)

Admins can configure the following in Settings → Global Settings:

- **Security Webhook URL**: HTTP endpoint to receive security notifications
  - Receives POST requests with JSON payload when threats detected
  - Payload format: `{"event": "security_risk_detected", "path": "...", "userId": 1, "signature": "...", "timestamp": "..."}`
  
- **Quarantine Path**: Directory path for quarantined files
  - Default: `/tmp/quarantine` if not configured
  - Files are moved here automatically when threats detected
  - Filenames include timestamp for tracking

### Docker Compose

The `compose.yaml` has been updated to include ClamAV service:

```yaml
services:
  filebrowser:
    environment:
      - CLAMAV_HOST=clamav
      - CLAMAV_PORT=3310
    depends_on:
      clamav:
        condition: service_healthy

  clamav:
    image: clamav/clamav:latest
    healthcheck:
      test: ["CMD", "/usr/local/bin/clamdcheck.sh"]
      interval: 60s
      timeout: 10s
      retries: 3
      start_period: 300s
```

## Usage

### Running with Docker Compose

1. Start the services:
   ```bash
   docker compose up --build
   ```

2. Wait for ClamAV to initialize (this may take 5-10 minutes on first startup as it downloads virus definitions)

3. Access File Browser at `http://localhost:8000`

### Upload Flow

1. **Upload a File**: Use the upload button or drag-and-drop
2. **Monitor Status**: The upload dialog shows:
   - "Uploading" - File is being transferred
   - "Security Scanning" - ClamAV is scanning the file
   - "Uploaded" (green checkmark) - File is clean and ready
   - "Security Risk" (red warning) - Threat detected

### Admin Security Management

Admins can access the Security Risks dashboard:

1. Navigate to **Settings** → **Security Risks**
2. View all files flagged as security threats
3. See threat details including:
   - File name and path
   - User who uploaded the file
   - Threat signature detected
   - Scan timestamp
   - Current status (Security Risk or Overridden)
4. Take action on each file:
   - **Override** (shield icon): Mark as false positive - file becomes available to all users
   - **Quarantine** (archive icon): Move to quarantine folder for isolation
   - **Delete** (trash icon): Permanently remove the file

### Admin Download Override

Admins can download security risk files for inspection:

1. Navigate to the file in the file browser
2. Add `?admin_override=true` to the download URL
3. Download will proceed despite security flag
4. Action is logged for audit purposes

## API Endpoints

### Get Scan Status
```http
GET /api/scan/status?path=/path/to/file
```

Returns the current scan status of a file.

**Response:**
```json
{
  "path": "/path/to/file",
  "userId": 1,
  "status": "clean",
  "scannedAt": "2024-01-01T12:00:00Z",
  "uploadedAt": "2024-01-01T12:00:00Z"
}
```

Status values:
- `uploading` - File is being uploaded
- `scanning` - Currently being scanned
- `clean` - No threats detected
- `security_risk` - Threat detected
- `scan_error` - Error during scanning
- `overridden` - Admin marked as false positive

### List Security Risks (Admin Only)
```http
GET /api/scan/risks
```

Returns all files flagged as security risks.

### Delete Security Risk (Admin Only)
```http
DELETE /api/scan/risks?path=/full/path/to/file&action=delete
```

Permanently deletes a file that has been flagged as a security risk.

**Query Parameters:**
- `path`: Full system path to the file
- `action`: `delete` (permanent removal) or `quarantine` (move to quarantine)

### Override Security Risk (Admin Only)
```http
POST /api/scan/override?path=/full/path/to/file
```

Marks a security risk as a false positive. File becomes available to all users but admins can still see it was flagged.

## Testing

### Testing with EICAR Test File

The EICAR test file is a safe way to test virus detection without using actual malware:

1. Create a text file with this exact content:
   ```
   X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*
   ```

2. Upload this file through File Browser

3. The file should be flagged as "EICAR-Test-File" (or similar) by ClamAV

4. Verify:
   - Upload shows "Security Risk" status
   - Download is blocked
   - File appears in Admin Security Risks dashboard

### Testing Without ClamAV

To test graceful degradation:

1. Stop the ClamAV container:
   ```bash
   docker compose stop clamav
   ```

2. Upload files - they should upload normally without scanning

3. Check logs - should show "ClamAV scanner is not available"

### Manual Testing Checklist

- [ ] Upload a clean file - should show "Uploaded" with green checkmark
- [ ] Upload EICAR test file - should show "Security Risk" with red warning
- [ ] Try to download security risk file - should be blocked with 403 error
- [ ] Admin downloads with `?admin_override=true` - should succeed
- [ ] View Security Risks as admin - EICAR file should be listed
- [ ] Override security risk - should mark as "False Positive"
- [ ] Download overridden file as normal user - should work
- [ ] Quarantine security risk file - should move to quarantine folder
- [ ] Delete security risk file from admin panel - should be removed
- [ ] Upload file with ClamAV stopped - should work without scanning
- [ ] Upload multiple files - scan status should update for each
- [ ] Configure webhook URL - should receive notification on threat detection

## Troubleshooting

### ClamAV Not Starting

**Issue**: ClamAV container keeps restarting or fails health check

**Solutions**:
- Increase `start_period` in health check (ClamAV needs time to download definitions)
- Check logs: `docker compose logs clamav`
- Ensure sufficient memory (ClamAV requires at least 1GB RAM)

### Files Not Being Scanned

**Issue**: All files show "Uploaded" immediately without scanning

**Check**:
1. ClamAV container is running: `docker compose ps`
2. Environment variables are set correctly
3. Backend logs show scanner availability

### False Positives

**Issue**: Legitimate files flagged as threats

**Actions**:
1. Review threat signature in Security Risks dashboard
2. Use the **Override** button (shield icon) to mark as false positive
3. File will be marked as "overridden" and available for download by all users
4. Admins can still see it's a flagged file in the dashboard
5. Alternatively, use the **Quarantine** button to isolate the file for further review

## Architecture

### Scan Flow

```
User uploads file
     ↓
File saved to disk
     ↓
Upload completes
     ↓
Backend triggers async scan
     ↓
ClamAV scans file
     ↓
Status stored in memory
     ↓
Frontend polls for status
     ↓
Display result to user
```

### Components

**Backend**:
- `scanner/scanner.go` - ClamAV integration
- `scanner/storage.go` - Scan metadata storage
- `http/scanner.go` - API handlers
- `http/tus_handlers.go` - Upload flow integration
- `http/raw.go` - Download protection

**Frontend**:
- `src/api/scanner.ts` - API client
- `src/stores/upload.ts` - Upload state management
- `src/components/prompts/UploadFiles.vue` - Upload UI
- `src/views/settings/SecurityRisks.vue` - Admin dashboard

## Security Considerations

1. **Defense in Depth**: Virus scanning is one layer - should be combined with other security measures
2. **Signature Updates**: ClamAV automatically updates virus definitions
3. **Zero-Day Threats**: ClamAV may not detect brand new malware
4. **Performance**: Scanning large files may take time - asynchronous design prevents blocking
5. **Storage**: Scan metadata is stored in memory - restart clears history

## Future Enhancements

Potential improvements for future versions:

- Persistent storage for scan history (database integration)
- Scan progress percentage for large files
- Configurable scan policies (skip certain file types, size limits)
- Integration with other scanning engines
- Scan file on download in addition to upload
- API rate limiting for scan endpoints
- Scheduled scans of existing files
- Detailed audit log of admin actions

## License

This feature follows the same license as File Browser (Apache License 2.0).
