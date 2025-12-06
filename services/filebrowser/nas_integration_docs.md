# 🗂️ NAS Integration Architecture — Electron App + Backend + NAS Microservice  
### (Full Auth + File Open Flow Documentation)

## 1. Overview
This document defines how the **Electron desktop application**, **backend**, and the **NAS microservice** will communicate so that:

- Users **log in once** (to your Electron app).
- Backend **creates and manages NAS auth tokens**.
- Electron can **download/open/save files locally** without requiring NAS login.
- Desktop apps (Photoshop, Illustrator, DWG viewers, etc.) can open these files as **local OS files**.

## 2. High-Level Architecture
```
Electron App (Renderer + Main)
           |
           | JWT (user session)
           v
Backend (Auth + File Access + NAS token management)
           |
           | Internal request
           v
NAS Microservice (running on NAS)
           |
           | NAS access (qtoken/sid, SMB/Web APIs)
           v
NAS Storage (files/folders)
```

## 3. Components

### 3.1 Electron App
- Renderer:
  - Displays UI (your current web portal).
  - Sends file operations to main via IPC.
  - Maintains user auth (accessToken/refreshToken).
- Main:
  - Calls backend for file downloads/uploads.
  - Saves downloaded file to local temp folder.
  - Opens file in OS default app using `shell.openPath()`
  - (Optional) Watches file for modification → uploads back.

### 3.2 Backend
Responsibilities:
- Authenticate users (`/auth/login`).
- Issue & refresh user access tokens (JWT).
- Maintain NAS sessions for the microservice.
- Validate user permissions.
- Forward file operations to the NAS microservice.
- Stream files back to Electron.

### 3.3 NAS Microservice
Responsibilities:
- Acts as a privileged proxy between backend ↔ NAS.
- Has direct NAS access without user login.
- Implements operations: list, download, upload, delete, rename.
- Uses NAS APIs internally.

## 4. Full Auth Flow

### Step 1: User logs into Electron App
```
POST /auth/login
{
  "email": "...",
  "password": "..."
}
```

### Step 2: Backend prepares NAS token
- Checks cached NAS session.
- If expired → refresh via NAS microservice.
- Electron never sees NAS creds.

## 5. File Open Flow

### Renderer → Main
```
ipcRenderer.send("open-remote-file", { fileId });
```

### Main → Backend
```
GET /files/:id/download
Authorization: Bearer <accessToken>
```

### Backend
- Validates JWT & permissions.
- Ensures NAS session.
- Streams file bytes from NAS microservice.

### Main (Electron)
- Saves file to temp folder.
- Calls `shell.openPath(localPath)`.

## 6. Optional Save-Back
- Watch temp file.
- On modify/close → upload:
```
PUT /files/:id/upload
```

## 7. API Contract
- `POST /auth/login`
- `POST /auth/refresh`
- `GET /files/:id/download`
- `PUT /files/:id/upload`
- Internal: `/nas/*`

## 8. Electron IPC
Renderer → Main:
```
open-remote-file { fileId }
```

Main → Renderer:
```
open-progress / open-error / open-success
```

## 9. Benefits
- No NAS login per machine
- Centralized permissions
- Secure: clients never see NAS tokens
- Smooth file opening workflow

## 10. Summary Workflow
```
User → Renderer → Main → Backend → NAS Microservice → NAS
                     ↓
                 Temp Local Path → OS App → (optional) Upload
```
