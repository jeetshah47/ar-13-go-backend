# QNAP Service Account Setup Guide

This guide explains how to set up a service account on your QNAP NAS for use with the AR-13 backend application.

## Overview

The service account is used by the backend to authenticate with the QNAP NAS File Station API. This allows the Electron app to access files on the NAS without requiring individual user logins.

## Prerequisites

- Access to QNAP NAS web interface (QTS)
- Administrator privileges on the NAS
- Network access to the NAS from the backend server

## Step 1: Access QNAP Control Panel

1. Open a web browser and navigate to your QNAP NAS IP address (e.g., `http://192.168.1.100`)
2. Log in with an administrator account

## Step 2: Create Service Account

1. Go to **Control Panel** → **Privilege** → **Users**
2. Click **Create** → **Create a user**
3. Fill in the user details:
   - **User name**: `ar13_service` (or your preferred name)
   - **Password**: Create a strong password (save this securely)
   - **Email**: Optional (can be left empty)
   - **Description**: "Service account for AR-13 backend application"
4. Click **Next**

## Step 3: Configure Permissions

1. **Shared Folder Permissions**:
   - Select the shared folder(s) that contain your files (e.g., "studio-work")
   - Set permissions to **Read/Write** for the service account
   - Click **Next**

2. **Application Access**:
   - Enable **File Station** access
   - Enable **SMB/CIFS** access (for network mounting)
   - Click **Next**

3. **Quota** (Optional):
   - Set quota if needed, or leave unlimited
   - Click **Next**

4. **User Groups** (Optional):
   - Add to groups if needed, or skip
   - Click **Next**

5. Review settings and click **Finish**

## Step 4: Enable File Station API

1. Go to **Control Panel** → **Network Services** → **File Station**
2. Ensure **File Station** is enabled
3. Check **Enable File Station API** (if available in your QTS version)
4. Note the API port (default: 8080)

## Step 5: Configure Backend Environment Variables

Add the following environment variables to your backend configuration:

```bash
# QNAP NAS Configuration
QNAP_NAS_IP=192.168.1.100              # Your QNAP NAS IP address
QNAP_API_PORT=8080                     # QNAP API port (default: 8080)
QNAP_SERVICE_USER=ar13_service         # Service account username
QNAP_SERVICE_PASSWORD=your_password    # Service account password
QNAP_SHARE_NAME=studio-work            # Share name to access
QNAP_SESSION_TIMEOUT=30                # Session timeout in minutes (default: 30)
```

## Step 6: Test Connection

1. Start your backend server
2. Check logs for QNAP authentication messages
3. Test the mount credentials endpoint:
   ```bash
   curl -H "Authorization: Bearer <your_jwt_token>" \
        http://localhost:3000/api/nas/mount-credentials
   ```

## Security Best Practices

1. **Strong Password**: Use a strong, unique password for the service account
2. **Minimal Permissions**: Only grant access to folders that are needed
3. **Network Security**: Ensure the backend server and NAS are on a secure network
4. **Regular Rotation**: Consider rotating the service account password periodically
5. **Monitoring**: Monitor service account activity in QNAP logs
6. **IP Whitelisting**: If possible, configure IP whitelisting for API access

## Troubleshooting

### Authentication Fails

- Verify the service account username and password are correct
- Check that File Station API is enabled
- Ensure the NAS IP and port are correct
- Check network connectivity between backend and NAS

### Permission Denied

- Verify the service account has Read/Write permissions on the target folder
- Check that SMB/CIFS access is enabled for the service account
- Ensure the share name matches exactly (case-sensitive)

### Session Expires Quickly

- Adjust `QNAP_SESSION_TIMEOUT` if needed
- Check QNAP session timeout settings in Control Panel
- The backend automatically refreshes sessions before expiration

## Additional Resources

- [QNAP File Station API Documentation](https://download.qnap.com/dev/QNAP_QTS_File_Station_API_v4.1.pdf)
- [QNAP User Management Guide](https://www.qnap.com/en/how-to/tutorial/article/how-to-manage-users-and-groups-in-qts)

## Notes

- The service account password is stored in environment variables - never commit to version control
- For production, consider using a secrets management system
- The session ID (SID) is automatically managed by the backend and expires after 30 minutes (configurable)

