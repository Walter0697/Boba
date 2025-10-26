# Testing Guide for Release Pipeline

This document describes how to test the automated release pipeline for BOBA.

## Overview

The release pipeline consists of three main workflows:
1. **CI Workflow** - Continuous integration testing
2. **Release Workflow** - Development releases on main/master
3. **Tagged Release Workflow** - Production releases on version tags

## Testing CI Workflow

### Trigger
The CI workflow runs automatically on:
- Pull requests to main/master
- Pushes to main/master

### Manual Testing
1. Create a test branch:
   ```bash
   git checkout -b test/ci-workflow
   ```

2. Make a small change (e.g., update README):
   ```bash
   echo "# Test" >> test.md
   git add test.md
   git commit -m "test: CI workflow"
   git push origin test/ci-workflow
   ```

3. Create a pull request on GitHub

4. Verify the CI workflow runs:
   - Go to Actions tab
   - Check "CI" workflow is running
   - Verify tests pass on all platforms (Linux, macOS, Windows)
   - Check linting completes successfully
   - Verify build succeeds

### Expected Results
- ✅ Tests pass on all platforms
- ✅ Linting completes (may have warnings)
- ✅ Build succeeds
- ✅ Coverage report generated

## Testing Development Release Workflow

### Trigger
The release workflow runs on pushes to main/master (without tags).

### Manual Testing
1. Merge a PR or push directly to main:
   ```bash
   git checkout main
   git pull origin main
   git merge test/ci-workflow
   git push origin main
   ```

2. Monitor the workflow:
   - Go to Actions tab
   - Check "Release" workflow is running
   - Wait for completion (5-10 minutes)

3. Verify the release:
   - Go to Releases page
   - Check for a pre-release with version `v0.0.0-dev-<commit-hash>`
   - Verify all platform binaries are attached:
     - boba-linux-amd64.tar.gz
     - boba-linux-arm64.tar.gz
     - boba-darwin-amd64.tar.gz
     - boba-darwin-arm64.tar.gz
     - boba-windows-amd64.zip

4. Test a binary:
   ```bash
   # Download and test Linux binary
   wget https://github.com/YOUR_USERNAME/boba/releases/download/v0.0.0-dev-abc123/boba-linux-amd64.tar.gz
   tar -xzf boba-linux-amd64.tar.gz
   ./boba-linux-amd64 --version
   ```

### Expected Results
- ✅ Pre-release created with dev version
- ✅ All platform binaries present
- ✅ Release notes generated
- ✅ Binary runs and shows version info

## Testing Tagged Release Workflow

### Trigger
The tagged release workflow runs when pushing version tags.

### Manual Testing

#### 1. Create a Test Tag
```bash
# Ensure you're on main with latest changes
git checkout main
git pull origin main

# Create a test tag (use a test version)
git tag -a v0.0.1-test -m "Test release pipeline"

# Push the tag
git push origin v0.0.1-test
```

#### 2. Monitor the Workflow
1. Go to Actions tab
2. Check "Tagged Release" workflow is running
3. Monitor each job:
   - Test job should complete first
   - Build-and-release job should follow
4. Wait for completion (5-10 minutes)

#### 3. Verify the Release
1. Go to Releases page
2. Check for release `v0.0.1-test`
3. Verify release is NOT marked as pre-release
4. Check all artifacts are present:
   - boba-linux-amd64.tar.gz
   - boba-linux-arm64.tar.gz
   - boba-darwin-amd64.tar.gz
   - boba-darwin-arm64.tar.gz
   - boba-windows-amd64.zip
   - checksums.txt

#### 4. Verify Checksums
```bash
# Download checksums file
wget https://github.com/YOUR_USERNAME/boba/releases/download/v0.0.1-test/checksums.txt

# Download a binary
wget https://github.com/YOUR_USERNAME/boba/releases/download/v0.0.1-test/boba-linux-amd64.tar.gz

# Verify checksum
sha256sum -c checksums.txt --ignore-missing
```

#### 5. Test Binary Functionality
```bash
# Extract and test
tar -xzf boba-linux-amd64.tar.gz
chmod +x boba-linux-amd64

# Check version
./boba-linux-amd64 --version
# Should show: boba version v0.0.1-test

# Check help
./boba-linux-amd64 --help

# Run interactively (if possible)
./boba-linux-amd64
```

#### 6. Verify Release Notes
1. Check release notes include:
   - Installation instructions for each platform
   - Changelog with commits since last release
   - Checksums section
   - Usage instructions
   - Link to full changelog

#### 7. Clean Up Test Release
```bash
# Delete the test tag locally
git tag -d v0.0.1-test

# Delete the test tag remotely
git push origin :refs/tags/v0.0.1-test

# Manually delete the release on GitHub (Releases page)
```

### Expected Results
- ✅ Official release created (not pre-release)
- ✅ All platform binaries present and working
- ✅ Checksums file present and valid
- ✅ Release notes comprehensive and formatted
- ✅ Version info embedded in binaries
- ✅ Binaries are executable and functional

## Testing Local Build Scripts

### Linux/macOS Build Script

```bash
# Make script executable
chmod +x deploy/scripts/build-all.sh

# Run the build script
./deploy/scripts/build-all.sh

# Verify output
ls -lh dist/

# Test a binary
./dist/boba-linux-amd64 --version
```

### Windows Build Script

```powershell
# Run the build script
.\deploy\scripts\build-all.ps1

# Verify output
Get-ChildItem dist\

# Test a binary
.\dist\boba-windows-amd64.exe --version
```

### Expected Results
- ✅ All platform binaries built successfully
- ✅ Archives created (tar.gz for Unix, zip for Windows)
- ✅ Checksums file generated
- ✅ Binaries are executable and show version info

## Testing Cross-Platform Compatibility

### Linux Testing
```bash
# Test on different distributions
docker run -it --rm -v $(pwd):/app ubuntu:latest /app/boba-linux-amd64 --version
docker run -it --rm -v $(pwd):/app alpine:latest /app/boba-linux-amd64 --version
```

### macOS Testing
```bash
# Test on Intel Mac
./boba-darwin-amd64 --version

# Test on Apple Silicon Mac
./boba-darwin-arm64 --version
```

### Windows Testing
```powershell
# Test on Windows
.\boba-windows-amd64.exe --version
```

## Troubleshooting Tests

### Workflow Fails at Test Stage
1. Check test output in Actions logs
2. Run tests locally: `go test -v ./...`
3. Fix failing tests and push changes
4. Re-run workflow

### Workflow Fails at Build Stage
1. Check build logs for compilation errors
2. Verify Go version compatibility
3. Test local build: `go build -v .`
4. Check for platform-specific issues

### Release Not Created
1. Verify workflow completed successfully
2. Check GitHub Actions permissions (needs `contents: write`)
3. Verify tag format matches trigger pattern
4. Check for rate limiting or API issues

### Binary Doesn't Run
1. Verify correct platform/architecture
2. Check file permissions: `chmod +x boba-*`
3. Verify binary is not corrupted: check checksums
4. Test with `--version` flag first

### Checksums Don't Match
1. Re-download the binary
2. Verify download completed successfully
3. Check for network issues during download
4. Compare file sizes with release page

## Automated Testing Checklist

Before creating a production release:

- [ ] All unit tests pass locally
- [ ] Integration tests pass (if any)
- [ ] CI workflow passes on latest main
- [ ] Test release created and verified (v0.0.1-test)
- [ ] All platform binaries tested
- [ ] Checksums verified
- [ ] Version info correct in binaries
- [ ] Release notes generated correctly
- [ ] Installation instructions work
- [ ] No security vulnerabilities in dependencies

## Performance Testing

### Build Time
Monitor build times to ensure they remain reasonable:
- Expected: 5-10 minutes for full multi-platform build
- If longer: Check for network issues or resource constraints

### Binary Size
Check binary sizes are reasonable:
```bash
ls -lh dist/
# Expected sizes:
# Linux/macOS: 15-25 MB (compressed: 5-8 MB)
# Windows: 15-25 MB (compressed: 5-8 MB)
```

### Startup Time
Test binary startup performance:
```bash
time ./boba-linux-amd64 --version
# Should be < 1 second
```

## Security Testing

### Checksum Verification
Always verify checksums before distributing:
```bash
sha256sum -c checksums.txt
```

### Binary Scanning
Consider scanning binaries for vulnerabilities:
```bash
# Example with trivy
trivy fs ./dist/
```

### Dependency Audit
Check for vulnerable dependencies:
```bash
go list -json -m all | nancy sleuth
```

## Continuous Monitoring

### After Release
1. Monitor GitHub Issues for installation problems
2. Check download statistics
3. Monitor for security advisories
4. Track user feedback

### Regular Maintenance
1. Update dependencies monthly
2. Test on new OS versions
3. Verify compatibility with new Go versions
4. Update GitHub Actions versions

## Documentation

Keep these documents updated:
- [ ] README.md - Installation instructions
- [ ] RELEASE.md - Release process
- [ ] TESTING.md - This document
- [ ] Workflow files - Comments and documentation
