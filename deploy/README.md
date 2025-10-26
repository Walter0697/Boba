# Deploy Directory

This directory contains all deployment and release-related files for BOBA.

## 📁 Directory Structure

```
deploy/
├── scripts/              # Build and deployment scripts
│   ├── build-all.sh     # Linux/macOS multi-platform build script
│   └── build-all.ps1    # Windows PowerShell build script
│
└── docs/                # Deployment documentation
    ├── RELEASE.md                      # Complete release process guide
    ├── TESTING.md                      # Testing procedures
    ├── RELEASE_QUICK_REFERENCE.md      # Quick command reference
    ├── PIPELINE_SUMMARY.md             # Pipeline overview
    └── IMPLEMENTATION_TASK_20.md       # Implementation summary
```

## 🚀 Quick Start

### Build Locally

**Linux/macOS:**
```bash
chmod +x deploy/scripts/build-all.sh
./deploy/scripts/build-all.sh
```

**Windows (PowerShell):**
```powershell
.\deploy\scripts\build-all.ps1
```

Binaries will be created in the `dist/` directory.

### Create a Release

```bash
# Create and push a version tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions will automatically build and release
```

## 📚 Documentation

- **[RELEASE.md](docs/RELEASE.md)** - Complete release process guide with step-by-step instructions
- **[TESTING.md](docs/TESTING.md)** - Testing procedures for the release pipeline
- **[RELEASE_QUICK_REFERENCE.md](docs/RELEASE_QUICK_REFERENCE.md)** - Quick command reference
- **[PIPELINE_SUMMARY.md](docs/PIPELINE_SUMMARY.md)** - Overview of the automated pipeline
- **[IMPLEMENTATION_TASK_20.md](docs/IMPLEMENTATION_TASK_20.md)** - Implementation details

## 🔧 Scripts

### build-all.sh (Linux/macOS)
Builds BOBA for all supported platforms:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

Creates compressed archives and generates SHA256 checksums.

### build-all.ps1 (Windows)
PowerShell version of the build script with the same functionality.

## 🎯 GitHub Actions Workflows

The automated release pipeline is configured in `.github/workflows/`:
- **ci.yml** - Continuous integration testing
- **release.yml** - Development releases
- **tag-release.yml** - Production releases

## 📦 Release Artifacts

Each release includes:
- `boba-linux-amd64.tar.gz` - Linux 64-bit
- `boba-linux-arm64.tar.gz` - Linux ARM 64-bit
- `boba-darwin-amd64.tar.gz` - macOS Intel
- `boba-darwin-arm64.tar.gz` - macOS Apple Silicon
- `boba-windows-amd64.zip` - Windows 64-bit
- `checksums.txt` - SHA256 checksums

## 🔗 Related Files

- `.github/workflows/` - GitHub Actions workflow definitions
- `main.go` - Contains version variables for build-time injection

## 💡 Tips

- Always test builds locally before creating a release
- Follow semantic versioning (MAJOR.MINOR.PATCH)
- Verify checksums after downloading release artifacts
- Check the Actions tab on GitHub to monitor workflow progress

## 🆘 Need Help?

- Check [RELEASE_QUICK_REFERENCE.md](docs/RELEASE_QUICK_REFERENCE.md) for quick commands
- Review [TESTING.md](docs/TESTING.md) for troubleshooting
- See [PIPELINE_SUMMARY.md](docs/PIPELINE_SUMMARY.md) for pipeline details
