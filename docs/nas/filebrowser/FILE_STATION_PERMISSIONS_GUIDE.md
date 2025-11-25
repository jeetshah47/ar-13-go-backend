# Complete Guide: Setting FileBrowser Permissions via File Station

This guide provides step-by-step instructions for setting up FileBrowser permissions using QNAP's File Station interface.

## Prerequisites

- Access to QNAP NAS web interface
- File Station application installed
- Admin or user with folder management permissions

---

## Step 1: Open File Station

1. **Log into your QNAP NAS** web interface
2. **Open File Station** from the main menu
3. You should see your shared folders listed

---

## Step 2: Navigate to Container Folder

1. **Locate the "Container" folder** in the shared folders list
2. **Double-click** on "Container" folder to open it
3. You should now be inside the Container folder

---

## Step 3: Create filebrowser-config Folder

### Method A: Using Right-Click Menu

1. **Right-click** in an empty area inside the Container folder
2. **Select "Create Folder"** or **"New Folder"** from the context menu
3. **Enter folder name:** `filebrowser-config`
4. **Click "Create"** or **"OK"**

### Method B: Using Toolbar

1. **Click the "New" button** in the File Station toolbar
2. **Select "Create Folder"**
3. **Enter folder name:** `filebrowser-config`
4. **Click "Create"** or **"OK"**

### Verification

- You should now see `filebrowser-config` folder in the Container directory
- If you don't see it, refresh the view (F5) or navigate back and into Container again

---

## Step 4: Set Folder Permissions

### Option 1: Using Properties Dialog (Recommended)

1. **Right-click** on the `filebrowser-config` folder you just created
2. **Select "Properties"** from the context menu
3. **Click on the "Permissions" tab** (or "Security" tab, depending on your QNAP version)

#### Set Permission Values:

4. **Click "Edit"** or **"Modify"** button (usually at the bottom or top of the permissions list)

5. **Set Owner Permissions:**
   - Find the **Owner** row (usually shows "admin" or your username)
   - Check these boxes:
     - ✅ **Read** (r)
     - ✅ **Write** (w)
     - ✅ **Execute** (x)
   - This equals: **rwx** (permission value 7)

6. **Set Group Permissions:**
   - Find the **Group** row (usually "administrators" or "users")
   - Check these boxes:
     - ✅ **Read** (r)
     - ✅ **Execute** (x)
     - ❌ **Write** (leave unchecked)
   - This equals: **r-x** (permission value 5)

7. **Set Others Permissions:**
   - Find the **Others** row (or "Everyone")
   - Check these boxes:
     - ✅ **Read** (r)
     - ✅ **Execute** (x)
     - ❌ **Write** (leave unchecked)
   - This equals: **r-x** (permission value 5)

8. **Apply Recursively:**
   - ✅ Check **"Apply to all subfolders and files"**
   - ✅ Check **"Apply changes to subfolders"** (if available)
   - This ensures permissions apply to all future files/folders

9. **Click "Apply"** or **"OK"** to save

### Option 2: Using Numeric Permissions (If Available)

Some QNAP versions allow direct numeric input:

1. **Right-click** → **Properties** → **Permissions**
2. **Look for "Permission" or "Mode" field**
3. **Enter:** `755`
4. **Check "Apply recursively"**
5. **Click "Apply"**

### Option 3: Using Security Settings

If "Permissions" tab is not available:

1. **Right-click** → **Properties** → **Security** tab
2. **Click "Edit"** or **"Advanced"**
3. **Modify permissions** as described in Option 1
4. **Apply recursively**

---

## Step 5: Set Folder Ownership

### Method 1: Via Properties Dialog

1. **Right-click** on `filebrowser-config` folder
2. **Select "Properties"**
3. **Look for "Ownership" tab** or **"Owner" field** in Permissions tab

4. **Change Owner:**
   - Click **"Change"** or **"Edit"** next to Owner
   - Select user with **UID 1000** (usually `admin`)
   - If unsure, select **"admin"** user
   - Click **"OK"**

5. **Apply Recursively:**
   - ✅ Check **"Apply to all subfolders and files"**
   - Click **"Apply"**

### Method 2: Via Control Panel

1. **Open Control Panel** from QNAP main menu
2. **Go to "Privilege Settings"** → **"Shared Folders"**
3. **Find "Container"** in the list
4. **Click "Edit"** (pencil icon)
5. **Go to "Permissions" tab**
6. **Find "filebrowser-config"** subfolder or set permissions for Container folder
7. **Set owner** to `admin` (UID 1000)
8. **Set permissions** to 755
9. **Apply changes**

---

## Step 6: Verify Permissions

### Check via File Station:

1. **Right-click** on `filebrowser-config` folder
2. **Select "Properties"** → **"Permissions" tab**
3. **Verify:**
   - Owner: `admin` (or user with UID 1000)
   - Permissions show: `rwxr-xr-x` or `755`
   - All checkboxes are set correctly

### Check via Command Line (Optional):

If you have SSH/ttyd access:

```bash
ls -la /share/Container/ | grep filebrowser-config
```

Should show:
```
drwxr-xr-x 2 admin administrators 4096 Nov 22 10:00 filebrowser-config
```

Where:
- `drwxr-xr-x` = permissions 755
- `admin` = owner
- `administrators` = group

---

## Step 7: Common Permission Values Reference

| Permission | Numeric | Binary | Description |
|------------|---------|--------|-------------|
| Read | 4 | 100 | Can view/list files |
| Write | 2 | 010 | Can create/modify files |
| Execute | 1 | 001 | Can enter/run directory |
| Read + Write | 6 | 110 | Can view and modify |
| Read + Execute | 5 | 101 | Can view and enter |
| Read + Write + Execute | 7 | 111 | Full access |

### Common Combinations:

- **755** = Owner: rwx (7), Group: r-x (5), Others: r-x (5)
- **777** = Everyone has full access (not recommended)
- **700** = Only owner has access

**For FileBrowser, use: 755**

---

## Troubleshooting

### Issue: Can't Find Permissions Option

**Solution:**
- Look for **"Security"** tab instead of "Permissions"
- Try **Control Panel** → **Privilege Settings** → **Shared Folders**
- Some QNAP models have permissions under **"Share Properties"**

### Issue: Can't Change Owner

**Solution:**
- You need **admin** privileges
- Try using **Control Panel** → **Users** to check user permissions
- Or use command line: `chown -R 1000:1000 /share/Container/filebrowser-config`

### Issue: Permissions Don't Apply Recursively

**Solution:**
- Some File Station versions don't support recursive permissions
- Use command line instead:
  ```bash
  chmod -R 755 /share/Container/filebrowser-config
  chown -R 1000:1000 /share/Container/filebrowser-config
  ```

### Issue: Don't Know Which User Has UID 1000

**Solution:**
1. **Control Panel** → **Users**
2. **Check user list** - usually `admin` has UID 1000
3. **Or use command line:**
   ```bash
   id admin
   # Should show: uid=1000(admin) gid=100(administrators)
   ```

### Issue: Folder Already Exists with Wrong Permissions

**Solution:**
1. **Delete the folder** (right-click → Delete)
2. **Recreate it** following Step 3
3. **Set permissions** following Step 4

Or fix existing folder:
1. **Right-click** → **Properties** → **Permissions**
2. **Click "Edit"**
3. **Set correct permissions** (755)
4. **Apply recursively**

---

## Visual Guide (What to Look For)

### Properties Dialog Should Show:

```
Properties: filebrowser-config
├── General Tab
│   ├── Name: filebrowser-config
│   ├── Location: /share/Container/
│   └── Size: (will show after files are created)
│
├── Permissions Tab ← YOU ARE HERE
│   ├── Owner: admin (UID: 1000)
│   │   ├── Read: ✅
│   │   ├── Write: ✅
│   │   └── Execute: ✅
│   │
│   ├── Group: administrators
│   │   ├── Read: ✅
│   │   ├── Write: ❌
│   │   └── Execute: ✅
│   │
│   └── Others: 
│       ├── Read: ✅
│       ├── Write: ❌
│       └── Execute: ✅
│
└── [Apply to all subfolders and files] ✅
```

---

## Quick Checklist

Before starting FileBrowser container, verify:

- [ ] `filebrowser-config` folder exists in `/share/Container/`
- [ ] Folder permissions are set to **755** (rwxr-xr-x)
- [ ] Owner is set to **admin** (or user with UID 1000)
- [ ] Permissions are applied recursively
- [ ] You can see the folder in File Station

---

## After Setting Permissions

1. **Start FileBrowser:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml up -d
   ```

2. **Check if it's working:**
   ```bash
   docker logs filebrowser
   ```

3. **Access FileBrowser:**
   - http://YOUR-NAS-IP:8081
   - Login with: `admin` / `admin` (or check logs for password)

---

## Alternative: Use Named Volume (No Permission Issues)

If you continue having permission issues, the docker-compose file is already configured to use a **named volume** which doesn't require manual permission setup. Just run:

```bash
docker-compose -f docker-compose.filebrowser.nas.yml up -d
```

The named volume approach automatically handles permissions!

---

## Need Help?

- Check `FIX_PERMISSIONS_FINAL.md` for troubleshooting
- Use the setup script: `./setup-filebrowser.sh`
- Check container logs: `docker logs filebrowser`

