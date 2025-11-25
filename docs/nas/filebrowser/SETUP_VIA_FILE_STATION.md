# Setting Up FileBrowser Using File Station

You can use QNAP's File Station to create directories and set permissions without using command line.

## Step 1: Create the Config Directory

1. **Open File Station** on your QNAP NAS
2. **Navigate to** `/share/Container/` folder
3. **Create new folder:**
   - Right-click in the empty space
   - Select **"Create Folder"** or **"New Folder"**
   - Name it: `filebrowser-config`
   - Click **"Create"** or **"OK"**

## Step 2: Set Permissions via File Station

### Method 1: Using Properties Dialog (Detailed Steps)

1. **Right-click** on the `filebrowser-config` folder you just created
2. **Select "Properties"** from the context menu
   - A dialog window will open showing folder properties

3. **Click on the "Permissions" tab**
   - If you don't see "Permissions" tab, look for **"Security"** tab instead
   - Different QNAP versions may have different tab names

4. **Click "Edit" or "Modify" button**
   - Usually located at the bottom or top of the permissions list
   - This allows you to change permissions

5. **Set Owner Permissions:**
   - Find the row showing **"Owner"** (usually shows "admin" or your username)
   - Check these three boxes:
     - ✅ **Read** (allows viewing files)
     - ✅ **Write** (allows creating/modifying files)
     - ✅ **Execute** (allows entering the directory)
   - This gives Owner: **rwx** (permission value = 7)

6. **Set Group Permissions:**
   - Find the row showing **"Group"** (usually "administrators" or "users")
   - Check these boxes:
     - ✅ **Read** (check)
     - ❌ **Write** (leave unchecked)
     - ✅ **Execute** (check)
   - This gives Group: **r-x** (permission value = 5)

7. **Set Others Permissions:**
   - Find the row showing **"Others"** or **"Everyone"**
   - Check these boxes:
     - ✅ **Read** (check)
     - ❌ **Write** (leave unchecked)
     - ✅ **Execute** (check)
   - This gives Others: **r-x** (permission value = 5)

8. **Final Result:**
   - Owner: rwx (7)
   - Group: r-x (5)
   - Others: r-x (5)
   - **Total: 755**

9. **Apply Recursively:**
   - ✅ Check the box: **"Apply to all subfolders and files"**
   - ✅ Check: **"Apply changes to subfolders"** (if available)
   - This ensures permissions apply to all future files/folders created inside

10. **Save Changes:**
    - Click **"Apply"** button
    - Then click **"OK"** to close the dialog
    - Wait for the operation to complete

### Method 2: Using Security Settings

1. **Right-click** on `filebrowser-config` folder
2. **Select "Security"** or **"Share Properties"**
3. **Click "Edit"** or **"Modify Permissions"**
4. **Add/Edit permissions:**
   - Ensure the owner has **Full Control** or **Read/Write/Execute**
   - Set permissions to **755** (rwxr-xr-x)

### Method 3: Using Control Panel (Alternative)

1. **Open Control Panel** on your NAS
2. **Go to "Privilege Settings"** → **"Shared Folders"**
3. **Find or create** `Container` folder
4. **Click "Edit"** → **"Permissions"**
5. **Set permissions** for the `filebrowser-config` subfolder

## Step 3: Verify Permissions

1. **Right-click** on `filebrowser-config` folder
2. **Select "Properties"**
3. **Check the "Permissions" tab:**
   - Should show: `rwxr-xr-x` or `755`
   - Owner should be `admin` or user with UID 1000

## Step 4: Set Ownership (If Available)

If File Station shows ownership options:

1. **Right-click** on `filebrowser-config` folder
2. **Select "Properties"** → **"Ownership"** tab
3. **Change owner to:**
   - `admin` (usually UID 1000)
   - Or `httpdusr` (web server user, often UID 1000)
   - Or look for user with ID `1000`

4. **Apply recursively** if option is available

## Alternative: Find the Correct User

If you're not sure which user has UID 1000:

1. **Open Control Panel** → **"Users"**
2. **Look for users** and check their UID
3. **Common users with UID 1000:**
   - `admin`
   - `httpdusr`
   - `httpd` (web server user)

## Quick Setup Checklist

- [ ] Created `/share/Container/filebrowser-config` folder via File Station
- [ ] Set permissions to **755** (rwxr-xr-x)
- [ ] Set owner to user with **UID 1000** (usually `admin`)
- [ ] Applied permissions recursively to subfolders
- [ ] Verified permissions in Properties dialog

## After Setup

Once permissions are set:

1. **Start FileBrowser container:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml up -d
   ```

2. **Uncomment the user line** in `docker-compose.filebrowser.nas.yml`:
   ```yaml
   user: "1000:1000"  # Uncomment this line
   ```

3. **Restart container:**
   ```bash
   docker-compose -f docker-compose.filebrowser.nas.yml restart
   ```

## Troubleshooting

### Can't Find Permissions Option
- Some QNAP models have permissions under **"Properties"** → **"Security"**
- Or try **Control Panel** → **"Privilege Settings"**

### Can't Change Owner
- You may need **admin** privileges
- Try using **Control Panel** → **"Users"** → **"Edit"** permissions
- Or use command line: `chown -R 1000:1000 /share/Container/filebrowser-config`

### Permissions Don't Apply Recursively
- Manually set permissions on the folder
- File Station may not have recursive option - use command line:
  ```bash
  chmod -R 755 /share/Container/filebrowser-config
  ```

### Still Getting Permission Errors
- Verify owner is correct: Check folder properties
- Try running container without `user:` restriction first
- Check container logs: `docker logs filebrowser`

## Notes

- **File Station** provides a GUI way to manage permissions
- **Command line** (`chmod`, `chown`) is more reliable for recursive operations
- **UID 1000** is the default user ID for many container applications
- **755 permissions** = Owner: rwx (7), Group: rx (5), Others: rx (5)

