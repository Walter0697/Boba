# Deploy Directory Structure

All deployment and release-related files have been organized into the `/deploy` directory to keep the project root clean.

## 📁 New Structure

```
deploy/
├── README.md                           # Deploy directory overview
├── STRUCTURE.md                        # This file
│
├── scripts/                            # Build and deployment scripts
│   ├── build-all.sh                   # Linux/macOS multi-platform build
│   └── build-all.ps1                  # Windows PowerShell build
│
└── docs/                              # Deployment documentation
    ├── RELEASE.md                     # Complete release process guide
    ├── TESTING.md                     # Testing procedures
    ├── RELEASE_QUICK_REFERENCE.md     # Quick command reference
    ├── PIPELINE_SUMMARY.md            # Pipeline overview
    └── IMPLEMENTATION_TASK_20.md      # Implementation summary
```

## 🔄 What Changed

### Files Moved

**From `scripts/` to `deploy/scripts/`:**
- `build-all.sh` → `deploy/scripts/build-all.sh`
- `build-all.ps1` → `deploy/scripts/build-all.ps1`

**From root and `.github/` to `deploy/docs/`:**
- `RELEASE.md` → `deploy/docs/RELEASE.md`
- `IMPLEMENTATION_TASK_20.md` → `deploy/docs/IMPLEMENTATION_TASK_20.md`
- `.github/TESTING.md` → `deploy/docs/TESTING.md`
- `.github/RELEASE_QUICK_REFERENCE.md` → `deploy/docs/RELEASE_QUICK_REFERENCE.md`
- `.github/PIPELINE_SUMMARY.md` → `deploy/docs/PIPELINE_SUMMARY.md`

### Files Updated

**README.md:**
- Updated script paths: `./scripts/build-all.sh` → `./deploy/scripts/build-all.sh`
- Updated documentation links: `RELEASE.md` → `deploy/docs/RELEASE.md`

**All documentation files in `deploy/docs/`:**
- Updated internal script references
- Updated cross-references between docs

### Files Unchanged

**GitHub Actions workflows remain in `.github/workflows/`:**
- `.github/workflows/ci.yml`
- `.github/workflows/release.yml`
- `.github/workflows/tag-release.yml`

These stay in `.github/workflows/` as required by GitHub Actions.

## 🎯 Benefits

### Cleaner Project Root
- Deployment files are now organized in one place
- Easier to find release-related documentation
- Reduced clutter in the root directory

### Better Organization
- Scripts are grouped together
- Documentation is centralized
- Clear separation of concerns

### Easier Maintenance
- All deployment files in one location
- Simpler to update and maintain
- Better for new contributors

## 🚀 Usage

### Building Locally

**Linux/macOS:**
```bash
chmod +x deploy/scripts/build-all.sh
./deploy/scripts/build-all.sh
```

**Windows:**
```powershell
.\deploy\scripts\build-all.ps1
```

### Reading Documentation

All deployment documentation is now in `deploy/docs/`:

```bash
# Quick reference
cat deploy/docs/RELEASE_QUICK_REFERENCE.md

# Full release guide
cat deploy/docs/RELEASE.md

# Testing procedures
cat deploy/docs/TESTING.md
```

### Creating Releases

The release process hasn't changed:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

GitHub Actions workflows automatically handle the rest.

## 📚 Documentation Index

| Document | Purpose | Location |
|----------|---------|----------|
| Deploy Overview | Introduction to deploy directory | `deploy/README.md` |
| Structure Guide | This document | `deploy/STRUCTURE.md` |
| Release Guide | Complete release process | `deploy/docs/RELEASE.md` |
| Testing Guide | Testing procedures | `deploy/docs/TESTING.md` |
| Quick Reference | Quick commands | `deploy/docs/RELEASE_QUICK_REFERENCE.md` |
| Pipeline Summary | Pipeline overview | `deploy/docs/PIPELINE_SUMMARY.md` |
| Implementation | Task 20 details | `deploy/docs/IMPLEMENTATION_TASK_20.md` |

## 🔗 Quick Links

- **Main README:** [README.md](../README.md)
- **Deploy README:** [deploy/README.md](README.md)
- **Release Guide:** [deploy/docs/RELEASE.md](docs/RELEASE.md)
- **Quick Reference:** [deploy/docs/RELEASE_QUICK_REFERENCE.md](docs/RELEASE_QUICK_REFERENCE.md)

## ✅ Verification

To verify the structure is correct:

```bash
# Check deploy directory exists
ls -la deploy/

# Check scripts are present
ls -la deploy/scripts/

# Check docs are present
ls -la deploy/docs/

# Verify scripts are executable (Linux/macOS)
test -x deploy/scripts/build-all.sh && echo "✓ build-all.sh is executable"
```

## 🎓 For New Contributors

If you're new to the project and need to work with releases:

1. **Start here:** Read `deploy/README.md`
2. **Quick commands:** Check `deploy/docs/RELEASE_QUICK_REFERENCE.md`
3. **Full guide:** Read `deploy/docs/RELEASE.md`
4. **Testing:** Review `deploy/docs/TESTING.md`

All deployment-related files are now in the `/deploy` directory!
