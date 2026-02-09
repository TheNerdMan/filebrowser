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
4. Delete security risk files directly from the dashboard

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

### List Security Risks (Admin Only)
```http
GET /api/scan/risks
```

Returns all files flagged as security risks.

### Delete Security Risk (Admin Only)
```http
DELETE /api/scan/risks?path=/full/path/to/file
```

Deletes a file that has been flagged as a security risk.

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
- [ ] View Security Risks as admin - EICAR file should be listed
- [ ] Delete security risk file from admin panel - should be removed
- [ ] Upload file with ClamAV stopped - should work without scanning
- [ ] Upload multiple files - scan status should update for each

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
2. If confirmed false positive, delete the file from admin panel
3. Consider updating ClamAV definitions: `docker compose restart clamav`

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
- Quarantine folder for security risks
- Email notifications for admins on threats detected
- Integration with other scanning engines
- Scan file on download in addition to upload
- API rate limiting for scan endpoints

## License

This feature follows the same license as File Browser (Apache License 2.0).
