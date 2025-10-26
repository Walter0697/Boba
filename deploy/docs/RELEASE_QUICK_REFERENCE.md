# Release Pipeline Quick Reference

Quick commands and workflows for BOBA releases.

## Quick Commands

### Create a New Release
```bash
# 1. Ensure you're on main with latest changes
git checkout main && git pull origin main

# 2. Create and push tag (replace X.Y.Z with version)
git tag -a vX.Y.Z -m "Release vX.Y.Z: Brief description"
git push origin vX.Y.Z

# 3. Monitor at: https://github.com/YOUR_USERNAME/boba/actions
```

### Delete a Release Tag
```bash
# Delete local tag
git tag -d vX.Y.Z

# Delete remote tag
git push origin :refs/tags/vX.Y.Z

# Manually delete release on GitHub Releases page
```

### Build Locally
```bash
# Linux/macOS
./deploy/scripts/build-all.sh

# Windows
.\deploy\scripts\build-all.ps1

# Output in: ./dist/
```

### Test a Release Binary
```bash
# Download
wget https://github.com/YOUR_USERNAME/boba/releases/download/vX.Y.Z/boba-linux-amd64.tar.gz

# Extract
tar -xzf boba-linux-amd64.tar.gz

# Test
./boba-linux-amd64 --version
./boba-linux-amd64 --help
```

### Verify Checksums
```bash
# Download checksums
wget https://github.com/YOUR_USERNAME/boba/releases/download/vX.Y.Z/checksums.txt

# Verify
sha256sum -c checksums.txt --ignore-missing
```

## Workflow Triggers

| Workflow | Trigger | Result |
|----------|---------|--------|
| CI | PR or push to main | Tests + lint + build |
| Release | Push to main (no tag) | Pre-release (dev version) |
| Tagged Release | Push tag `v*.*.*` | Official release |

## Version Numbering

Follow [Semantic Versioning](https://semver.org/):

```
vMAJOR.MINOR.PATCH

Examples:
v1.0.0 - Initial release
v1.1.0 - New features
v1.1.1 - Bug fixes
v2.0.0 - Breaking changes
```

## Release Checklist

### Before Release
- [ ] All tests pass: `go test ./...`
- [ ] Code builds: `go build .`
- [ ] Documentation updated
- [ ] CHANGELOG updated (if maintained)
- [ ] Version number decided

### Creating Release
- [ ] Tag created with correct version
- [ ] Tag pushed to GitHub
- [ ] Workflow started successfully

### After Release
- [ ] Workflow completed successfully
- [ ] Release appears on GitHub
- [ ] All artifacts present (5 binaries + checksums)
- [ ] Download and test at least one binary
- [ ] Verify checksums
- [ ] Announce release (if applicable)

## Platform Artifacts

Each release includes:

```
boba-linux-amd64.tar.gz      # Linux 64-bit
boba-linux-arm64.tar.gz      # Linux ARM 64-bit
boba-darwin-amd64.tar.gz     # macOS Intel
boba-darwin-arm64.tar.gz     # macOS Apple Silicon
boba-windows-amd64.zip       # Windows 64-bit
checksums.txt                # SHA256 checksums
```

## Common Issues

### "Workflow not found"
- Check workflow file syntax
- Verify file is in `.github/workflows/`
- Ensure file has `.yml` extension

### "Permission denied"
- Check repository settings → Actions → General
- Enable "Read and write permissions"

### "Tests failed"
- Run locally: `go test ./...`
- Fix failing tests
- Push changes before tagging

### "Build failed"
- Check Go version compatibility
- Verify dependencies: `go mod tidy`
- Test local build: `go build .`

### "Release not created"
- Verify tag format: `v*.*.*`
- Check workflow logs for errors
- Ensure GitHub token has permissions

## Useful Links

- [Actions Tab](https://github.com/YOUR_USERNAME/boba/actions)
- [Releases Page](https://github.com/YOUR_USERNAME/boba/releases)
- [Full Release Guide](RELEASE.md)
- [Testing Guide](TESTING.md)

## Emergency Rollback

If a release has critical issues:

```bash
# 1. Delete the tag
git tag -d vX.Y.Z
git push origin :refs/tags/vX.Y.Z

# 2. Delete release on GitHub (manually)

# 3. Fix the issue

# 4. Create new patch version
git tag -a vX.Y.Z+1 -m "Release vX.Y.Z+1: Fix critical issue"
git push origin vX.Y.Z+1
```

## Support

For issues with the release pipeline:
1. Check [TESTING.md](TESTING.md) for troubleshooting
2. Review workflow logs in Actions tab
3. Open an issue with workflow logs attached
