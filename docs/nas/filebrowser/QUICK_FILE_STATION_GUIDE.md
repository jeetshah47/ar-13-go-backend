# Quick File Station Permissions Guide

**Fast reference for setting FileBrowser permissions via File Station.**

## Quick Steps (5 Minutes)

### 1. Create Folder
```
File Station → Container folder → Right-click → Create Folder → "filebrowser-config"
```

### 2. Set Permissions
```
Right-click filebrowser-config → Properties → Permissions tab → Edit
```

### 3. Set Values
- **Owner:** ✅ Read, ✅ Write, ✅ Execute (rwx = 7)
- **Group:** ✅ Read, ❌ Write, ✅ Execute (r-x = 5)
- **Others:** ✅ Read, ❌ Write, ✅ Execute (r-x = 5)
- **Result:** 755

### 4. Apply Recursively
- ✅ Check "Apply to all subfolders and files"
- Click "Apply" → "OK"

### 5. Set Owner (if needed)
- Properties → Ownership tab → Change to "admin"
- Apply recursively

## Permission Values Cheat Sheet

| What | Owner | Group | Others | Numeric |
|------|-------|-------|--------|---------|
| **Read** | ✅ | ✅ | ✅ | 4 |
| **Write** | ✅ | ❌ | ❌ | 2 |
| **Execute** | ✅ | ✅ | ✅ | 1 |
| **Total** | rwx | r-x | r-x | **755** |

## Visual Checklist

When you're done, Properties → Permissions should show:

```
Owner (admin):
  [✓] Read    [✓] Write    [✓] Execute

Group (administrators):
  [✓] Read    [ ] Write    [✓] Execute

Others:
  [✓] Read    [ ] Write    [✓] Execute

[✓] Apply to all subfolders and files
```

## Troubleshooting

**Can't find Permissions?**
→ Look for "Security" tab instead

**Can't change Owner?**
→ Use Control Panel → Privilege Settings → Shared Folders

**Still having issues?**
→ Use named volume (already configured in docker-compose) - no manual permissions needed!

## After Setup

```bash
docker-compose -f docker-compose.filebrowser.nas.yml up -d
docker logs filebrowser
```

Access: http://YOUR-NAS-IP:8081

