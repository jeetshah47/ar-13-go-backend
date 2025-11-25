# Understanding "studio work" Folder Structure

## How It Works

Since we mounted `/share/studio work` directly as MinIO's data directory, here's how files are organized:

### File Explorer View

When you open "studio work" folder in File Explorer (via network share or File Station), you'll see:

```
studio work/
├── ar-13-uploads/          ← MinIO bucket (created by backend)
│   ├── file1.pdf           ← Files uploaded via backend
│   ├── file2.jpg
│   └── subfolder/
│       └── file3.docx
├── documents/              ← Another MinIO bucket (if you create it)
│   └── doc1.pdf
├── images/                 ← Another MinIO bucket (if you create it)
│   └── img1.png
├── my-folder/              ← Regular folder (not a MinIO bucket)
│   └── manual-file.txt     ← Files you add manually
└── .minio-config/          ← Hidden (MinIO internal)
```

## Important Points

### 1. MinIO Buckets = Folders

- Each **MinIO bucket** appears as a **folder** in "studio work"
- The bucket name becomes the folder name
- Files uploaded via MinIO API go into these bucket folders

### 2. MINIO_BUCKET Environment Variable

```env
MINIO_BUCKET=ar-13-uploads
```

This tells your backend:
- **Which bucket to use** for uploads/downloads
- The bucket `ar-13-uploads` will appear as folder `ar-13-uploads/` in "studio work"
- Your backend will store files in: `studio work/ar-13-uploads/`

### 3. You Can See Everything

In File Explorer, you'll see:
- ✅ **All MinIO buckets** (as folders)
- ✅ **All files** in those buckets
- ✅ **Any other folders/files** you add manually to "studio work"
- ❌ Hidden folders (`.minio-config`, `.minio-certs`) - won't show in normal view

### 4. Multiple Buckets

You can create multiple buckets, and each will appear as a folder:

- Create bucket `documents` → See `studio work/documents/`
- Create bucket `images` → See `studio work/images/`
- Create bucket `ar-13-uploads` → See `studio work/ar-13-uploads/`

## Example Scenarios

### Scenario 1: Backend Uploads File

1. Backend uploads `report.pdf` to bucket `ar-13-uploads`
2. File appears in: `studio work/ar-13-uploads/report.pdf`
3. Visible in File Explorer immediately

### Scenario 2: Manual File Addition

1. You manually copy `manual-doc.docx` to `studio work/my-folder/`
2. File is visible in File Explorer
3. **But:** Not accessible via MinIO API (it's not in a bucket)

### Scenario 3: Multiple Buckets

1. Backend uses bucket `ar-13-uploads` (from MINIO_BUCKET)
2. You create bucket `documents` via MinIO Console
3. Both folders visible in File Explorer:
   - `studio work/ar-13-uploads/`
   - `studio work/documents/`

## Best Practice

### Recommended Setup

1. **Use a specific bucket name** in `.env`:
   ```env
   MINIO_BUCKET=ar-13-uploads
   ```

2. **Create the bucket** in MinIO Console (or it auto-creates on first upload)

3. **All backend files** go to: `studio work/ar-13-uploads/`

4. **In File Explorer**, you'll see:
   - The `ar-13-uploads` folder with all your files
   - Any other buckets you create
   - Any other folders you add manually

## Summary

- ✅ **MINIO_BUCKET** = Bucket name (e.g., `ar-13-uploads`)
- ✅ **File Explorer** shows ALL folders in "studio work"
- ✅ **MinIO buckets** appear as folders
- ✅ **Backend** uses the specified bucket
- ✅ **You can see everything** in File Explorer

The bucket name in `MINIO_BUCKET` is just which bucket your backend uses - you'll still see ALL content of "studio work" in File Explorer!

